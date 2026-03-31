package service

import (
	"context"
	"fmt"
	"log"
	aiclient "techmemo/backend/ai/client"
	"techmemo/backend/ai/queue"
	"techmemo/backend/common/errors"
	"techmemo/backend/config"
	"techmemo/backend/dao"
	"techmemo/backend/handler/dto"
	"techmemo/backend/model"

	"github.com/google/uuid"
)

type AIService struct {
	aiDao    *dao.AIDao
	chatDao  *dao.ChatDao
	noteDao  *dao.NoteDao
	queue    queue.Queue
	aiClient aiclient.AIClient
}

func (a *AIService) GetNoteAIStatus(ctx context.Context, noteID int64) (dto.GetNoteAIStatusResp, error) {
	logs, err := a.aiDao.GetLogsByNoteID(ctx, noteID)
	if err != nil {
		return dto.GetNoteAIStatusResp{}, errors.InternalErr
	}

	if len(logs) == 0 {
		return dto.GetNoteAIStatusResp{
			NoteID: noteID,
			Status: "not_started",
		}, nil
	}

	progress := make(map[string]string)
	overallStatus := "completed"

	for _, logItem := range logs {
		progress[logItem.ProcessType] = logItem.Status

		switch logItem.Status {
		case "failed":
			overallStatus = "failed"
		case "processing":
			if overallStatus != "failed" {
				overallStatus = "processing"
			}
		case "pending":
			if overallStatus != "failed" && overallStatus != "processing" {
				overallStatus = "pending"
			}
		}
	}

	return dto.GetNoteAIStatusResp{
		NoteID:   noteID,
		Status:   overallStatus,
		Progress: progress,
	}, nil
}

// GetTaskStatus 按 taskID 查询任务状态，适用于全局思维导图等非笔记任务。
func (a *AIService) GetTaskStatus(ctx context.Context, taskID string) (dto.GetTaskStatusResp, error) {
	logs, err := a.aiDao.GetLogsByTaskID(ctx, taskID)
	if err != nil {
		return dto.GetTaskStatusResp{}, errors.InternalErr
	}
	if len(logs) == 0 {
		return dto.GetTaskStatusResp{TaskID: taskID, Status: "not_started"}, nil
	}

	status := "completed"
	for _, logItem := range logs {
		switch logItem.Status {
		case "failed":
			status = "failed"
		case "processing":
			if status != "failed" {
				status = "processing"
			}
		case "pending":
			if status != "failed" && status != "processing" {
				status = "pending"
			}
		}
	}
	return dto.GetTaskStatusResp{TaskID: taskID, Status: status}, nil
}

func (a *AIService) SetQueue(q queue.Queue) {
	a.queue = q
}

func (a *AIService) cleanNoteAIData(ctx context.Context, noteID int64) {
	_ = a.aiDao.DeleteEmbeddingsByNoteID(ctx, noteID)
	_ = a.aiDao.DeleteRelationsByNoteID(ctx, noteID)
	_ = a.aiDao.DeleteNoteRootNodesByNoteID(ctx, noteID)
	_ = a.aiDao.DeleteKnowledgePointsByNoteID(ctx, noteID)
}

func (a *AIService) SubmitTask(ctx context.Context, noteID int64) (string, error) {
	a.cleanNoteAIData(ctx, noteID)

	taskID := generateTaskID(noteID)

	err := a.aiDao.CreateAILog(ctx, dao.CreateAILogParams{
		SourceNoteID: noteID,
		TaskID:       taskID,
		TargetType:   "note",
		TargetID:     noteID,
		ProcessType:  "classify",
		ModelName:    config.AppConfig.AI.Chat.Model,
		Status:       "pending",
	})
	if err != nil {
		return "", errors.InternalErr
	}

	if err := a.queue.Publish(queue.AITask{TaskID: taskID}); err != nil {
		return "", errors.InternalErr
	}
	return taskID, nil
}

// SubmitGlobalMindMapTask 提交全局思维导图生成任务，由 API 层调用。
func (a *AIService) SubmitGlobalMindMapTask(ctx context.Context, userID int64) (string, error) {
	taskID := generateGlobalTaskID(userID)

	// 清除旧的 global 关系缓存，确保本次重新生成
	if err := a.aiDao.DeleteGlobalRelationsByUserID(ctx, userID); err != nil {
		return "", errors.InternalErr
	}

	if err := a.aiDao.CreateAILog(ctx, dao.CreateAILogParams{
		SourceNoteID: 0,
		TaskID:       taskID,
		TargetType:   "user",
		TargetID:     userID,
		ProcessType:  "global_mindmap",
		ModelName:    config.AppConfig.AI.Chat.Model,
		Status:       "pending",
	}); err != nil {
		return "", errors.InternalErr
	}

	if err := a.queue.Publish(queue.AITask{TaskID: taskID}); err != nil {
		return "", errors.InternalErr
	}
	return taskID, nil
}

// ProcessTask 处理一个 AI 任务，由 Worker 调用。
func (a *AIService) ProcessTask(ctx context.Context, task queue.AITask) {
	logs, err := a.aiDao.GetLogsByTaskID(ctx, task.TaskID)
	if err != nil {
		log.Println("获取任务日志失败:", err)
		return
	}

	for _, logItem := range logs {
		if logItem.Status != "pending" {
			continue
		}

		a.aiDao.UpdateStatus(ctx, logItem.ID, "processing")

		switch logItem.ProcessType {
		case "classify":
			a.handleClassify(ctx, logItem)
		case "extract":
			a.handleExtract(ctx, logItem)
		case "embedding":
			a.handleEmbedding(ctx, logItem)
		case "global_mindmap":
			a.handleGlobalMindMap(ctx, logItem)
		default:
			a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		}
	}
}

func (a *AIService) handleClassify(ctx context.Context, logItem *model.AiProcessLog) {
	note, err := a.noteDao.GetNoteByID(ctx, logItem.TargetID)
	if err != nil {
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("获取笔记失败: %v", err)
		return
	}

	noteType, err := a.aiClient.ClassifyNoteType(ctx, note.ContentMd)
	if err != nil {
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("分类笔记类型失败: %v", err)
		return
	}

	if err := a.noteDao.UpdateNote(ctx, logItem.TargetID, dao.UpdateNoteParams{NoteType: &noteType}); err != nil {
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("更新笔记类型失败: %v", err)
		return
	}

	switch noteType {
	case "knowledge":
		_ = a.aiDao.CreateAILog(ctx, dao.CreateAILogParams{
			SourceNoteID: logItem.SourceNoteID,
			TaskID:       logItem.TaskID,
			TargetType:   "note",
			TargetID:     logItem.TargetID,
			ProcessType:  "extract",
			ModelName:    a.aiClient.GetChatModelName(),
			Status:       "pending",
		})
		_ = a.aiDao.CreateAILog(ctx, dao.CreateAILogParams{
			SourceNoteID: logItem.SourceNoteID,
			TaskID:       logItem.TaskID,
			TargetType:   "note",
			TargetID:     logItem.TargetID,
			ProcessType:  "embedding",
			ModelName:    a.aiClient.GetEmbeddingModelName(),
			Status:       "pending",
		})
		_ = a.queue.Publish(queue.AITask{TaskID: logItem.TaskID})
	case "reference":
		_ = a.aiDao.CreateAILog(ctx, dao.CreateAILogParams{
			SourceNoteID: logItem.SourceNoteID,
			TaskID:       logItem.TaskID,
			TargetType:   "note",
			TargetID:     logItem.TargetID,
			ProcessType:  "embedding",
			ModelName:    a.aiClient.GetEmbeddingModelName(),
			Status:       "pending",
		})
		_ = a.queue.Publish(queue.AITask{TaskID: logItem.TaskID})
	default:
		// ignore 类型，直接完成
	}

	a.aiDao.UpdateStatus(ctx, logItem.ID, "completed")
	log.Printf("笔记分类完成，笔记ID: %d, 类型: %s", logItem.TargetID, noteType)
}

func (a *AIService) handleExtract(ctx context.Context, logItem *model.AiProcessLog) {
	note, err := a.noteDao.GetNoteByID(ctx, logItem.TargetID)
	if err != nil {
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("获取笔记失败: %v", err)
		return
	}

	knowledgePoints, err := a.aiClient.ExtractKnowledgePoints(ctx, note.ContentMd)
	if err != nil {
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("抽取知识点失败: %v", err)
		return
	}

	// 递归展开树形结构，收集所有节点和父子关系（ID 保存后回填）
	type nodeEntry struct {
		model  *model.KnowledgePoint
		parent *model.KnowledgePoint
	}
	var entries []nodeEntry
	var collect func(nodes []aiclient.KnowledgePoint, parent *model.KnowledgePoint)
	collect = func(nodes []aiclient.KnowledgePoint, parent *model.KnowledgePoint) {
		for _, kp := range nodes {
			m := &model.KnowledgePoint{
				UserID:          note.UserID,
				Name:            kp.Name,
				Description:     kp.Description,
				ImportanceScore: kp.ImportanceScore,
				SourceNoteID:    logItem.TargetID,
			}
			entries = append(entries, nodeEntry{model: m, parent: parent})
			collect(kp.Children, m)
		}
	}
	collect(knowledgePoints, nil)

	allModels := make([]*model.KnowledgePoint, len(entries))
	for i, e := range entries {
		allModels[i] = e.model
	}

	if err := a.aiDao.SaveKnowledgePoints(ctx, allModels); err != nil {
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("保存知识点失败: %v", err)
		return
	}

	// 知识点落库后 ID 已填充，构建父子关系
	var relations []*model.KnowledgeRelation
	for _, e := range entries {
		if e.parent != nil && e.parent.ID > 0 && e.model.ID > 0 {
			relations = append(relations, &model.KnowledgeRelation{
				FromKnowledgeID: e.parent.ID,
				ToKnowledgeID:   e.model.ID,
				RelationType:    "related",
			})
		}
	}
	if err := a.aiDao.SaveKnowledgeRelations(ctx, relations); err != nil {
		log.Printf("保存知识点关系失败（不影响主流程）: %v", err)
	}

	// 保存顶节点（parent == nil 的节点）到 note_root_node 表
	for _, e := range entries {
		if e.parent == nil && e.model.ID > 0 {
			_ = a.aiDao.SaveNoteRootNode(ctx, &model.NoteRootNode{
				NoteID:          logItem.TargetID,
				RootKnowledgeID: e.model.ID,
				Name:            e.model.Name,
				Description:     e.model.Description,
				ImportanceScore: e.model.ImportanceScore,
			})
		}
	}

	for _, kp := range allModels {
		_ = a.aiDao.CreateAILog(ctx, dao.CreateAILogParams{
			SourceNoteID: logItem.SourceNoteID,
			TaskID:       logItem.TaskID,
			TargetType:   "knowledge",
			TargetID:     kp.ID,
			ProcessType:  "embedding",
			ModelName:    a.aiClient.GetEmbeddingModelName(),
			Status:       "pending",
		})
		_ = a.queue.Publish(queue.AITask{TaskID: logItem.TaskID})
	}

	a.aiDao.UpdateStatus(ctx, logItem.ID, "completed")
	log.Printf("知识抽取完成，笔记ID: %d, 节点数: %d, 关系数: %d", logItem.TargetID, len(allModels), len(relations))
}

func (a *AIService) handleEmbedding(ctx context.Context, logItem *model.AiProcessLog) {
	var text string

	switch logItem.TargetType {
	case "note":
		note, err := a.noteDao.GetNoteByID(ctx, logItem.TargetID)
		if err != nil {
			a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
			log.Printf("获取笔记失败: %v", err)
			return
		}
		text = note.Title + "\n" + note.ContentMd
	case "knowledge":
		kp, err := a.aiDao.GetKnowledgePointByID(ctx, logItem.TargetID)
		if err != nil {
			a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
			log.Printf("获取知识点失败: %v", err)
			return
		}
		text = kp.Name + "\n" + kp.Description
	case "chat_message":
		msg, err := a.chatDao.GetMessageByID(ctx, logItem.TargetID)
		if err != nil {
			a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
			log.Printf("获取聊天消息失败: %v", err)
			return
		}
		text = msg.Content
	default:
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		return
	}

	embedding, err := a.aiClient.GetEmbedding(ctx, text)
	if err != nil {
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("获取向量失败: %v", err)
		return
	}

	if err := a.aiDao.SaveEmbedding(ctx, &model.EmbeddingCustom{
		TargetType: logItem.TargetType,
		TargetID:   logItem.TargetID,
		Vector:     embedding,
		ModelName:  a.aiClient.GetEmbeddingModelName(),
	}); err != nil {
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		log.Printf("保存向量失败: %v", err)
		return
	}

	a.aiDao.UpdateStatus(ctx, logItem.ID, "completed")
}

// GetMindMap 从 DB 读取知识点和关系，重建树形结构
func (a *AIService) GetMindMap(ctx context.Context, noteID int64) (dto.GetMindMapResp, error) {
	// 1. 获取该笔记下所有知识点
	kps, err := a.aiDao.GetKnowledgePointsByNoteID(ctx, noteID)
	if err != nil {
		return dto.GetMindMapResp{}, errors.InternalErr
	}

	// 2. 获取关系
	relations, err := a.aiDao.GetKnowledgeRelationsByNoteID(ctx, noteID)
	if err != nil {
		return dto.GetMindMapResp{}, errors.InternalErr
	}

	// 3. 构建 id -> node 映射
	nodeMap := make(map[int64]*dto.MindMapNode, len(kps))
	for _, kp := range kps {
		n := &dto.MindMapNode{
			ID:              kp.ID,
			Name:            kp.Name,
			Description:     kp.Description,
			ImportanceScore: kp.ImportanceScore,
			Children:        []*dto.MindMapNode{},
		}
		nodeMap[kp.ID] = n
	}

	// 4. 根据 related 关系挂载子节点，同时记录哪些是子节点
	childIDs := make(map[int64]bool)
	for _, r := range relations {
		if r.RelationType != "related" {
			continue
		}
		parent, ok := nodeMap[r.FromKnowledgeID]
		child, ok2 := nodeMap[r.ToKnowledgeID]
		if !ok || !ok2 {
			continue
		}
		parent.Children = append(parent.Children, child)
		childIDs[r.ToKnowledgeID] = true
	}

	// 5. 顶层节点 = 不是任何人的子节点
	var roots []*dto.MindMapNode
	for _, kp := range kps {
		if !childIDs[kp.ID] {
			roots = append(roots, nodeMap[kp.ID])
		}
	}

	return dto.GetMindMapResp{NoteID: noteID, Nodes: roots}, nil
}

// GetGlobalMindMap 纯读库：返回已生成的全局思维导图（顶节点 + global 关系）。
// 若尚未生成，请先调用 POST /ai/mindmap/global 触发生成任务。
func (a *AIService) GetGlobalMindMap(ctx context.Context, userID int64) (dto.GetGlobalMindMapResp, error) {
	rootNodes, err := a.aiDao.GetRootNodesByUserID(ctx, userID)
	if err != nil {
		return dto.GetGlobalMindMapResp{}, errors.InternalErr
	}

	nodes := make([]dto.GlobalMindMapNode, len(rootNodes))
	for i, n := range rootNodes {
		nodes[i] = dto.GlobalMindMapNode{
			ID:              n.RootKnowledgeID,
			NoteID:          n.NoteID,
			Name:            n.Name,
			Description:     n.Description,
			ImportanceScore: n.ImportanceScore,
		}
	}

	cachedRelations, err := a.aiDao.GetGlobalRelationsByUserID(ctx, userID)
	if err != nil {
		return dto.GetGlobalMindMapResp{}, errors.InternalErr
	}

	edges := make([]dto.GlobalMindMapEdge, len(cachedRelations))
	for i, r := range cachedRelations {
		label := r.RelationType
		if len(label) > 7 && label[:7] == "global:" {
			label = label[7:]
		}
		edges[i] = dto.GlobalMindMapEdge{
			FromID: r.FromKnowledgeID,
			ToID:   r.ToKnowledgeID,
			Label:  label,
		}
	}

	return dto.GetGlobalMindMapResp{Nodes: nodes, Edges: edges}, nil
}

// handleGlobalMindMap 由 Worker 调用，执行全局思维导图的 AI 生成并写库。
func (a *AIService) handleGlobalMindMap(ctx context.Context, logItem *model.AiProcessLog) {
	userID := logItem.TargetID
	rootNodes, err := a.aiDao.GetRootNodesByUserID(ctx, userID)
	if err != nil || len(rootNodes) < 2 {
		log.Printf("全局思维导图：用户 %d 顶节点不足，跳过生成", userID)
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		return
	}

	aiNodes := make([]aiclient.GlobalNode, len(rootNodes))
	for i, n := range rootNodes {
		aiNodes[i] = aiclient.GlobalNode{
			ID:          n.RootKnowledgeID,
			Name:        n.Name,
			Description: n.Description,
		}
	}

	aiRelations, err := a.aiClient.BuildGlobalMindMap(ctx, aiNodes)
	if err != nil {
		log.Printf("全局思维导图 AI 生成失败: %v", err)
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		return
	}

	var relations []*model.KnowledgeRelation
	for _, r := range aiRelations {
		relations = append(relations, &model.KnowledgeRelation{
			FromKnowledgeID: r.FromID,
			ToKnowledgeID:   r.ToID,
			RelationType:    "global:" + r.Label,
		})
	}
	if err := a.aiDao.SaveKnowledgeRelations(ctx, relations); err != nil {
		log.Printf("保存全局关系失败: %v", err)
		a.aiDao.UpdateStatus(ctx, logItem.ID, "failed")
		return
	}

	a.aiDao.UpdateStatus(ctx, logItem.ID, "completed")
	log.Printf("全局思维导图生成完成，用户 %d，关系数: %d", userID, len(relations))
}

func generateTaskID(id int64) string {
	return fmt.Sprintf("ai:%d:%s", id, uuid.NewString())
}

func generateGlobalTaskID(userID int64) string {
	return fmt.Sprintf("global:%d:%s", userID, uuid.NewString())
}

func NewAIService(aiDao *dao.AIDao, noteDao *dao.NoteDao, chatDao *dao.ChatDao, aiClient aiclient.AIClient) *AIService {
	return &AIService{
		aiDao:    aiDao,
		noteDao:  noteDao,
		chatDao:  chatDao,
		aiClient: aiClient,
	}
}
