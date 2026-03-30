# Markdown 测试文档

这是一个用于测试 Markdown 渲染的完整示例文档。

## 1. 标题层级

### 三级标题
#### 四级标题
##### 五级标题
###### 六级标题

## 2. 文本样式

**粗体文本** 和 *斜体文本* 以及 ***粗斜体***

~~删除线文本~~

这是一段普通文本，包含 `行内代码` 的示例。

## 3. 列表

### 无序列表
- 第一项
- 第二项
  - 嵌套项 2.1
  - 嵌套项 2.2
    - 更深层嵌套 2.2.1
- 第三项

### 有序列表
1. 第一步
2. 第二步
   1. 子步骤 2.1
   2. 子步骤 2.2
3. 第三步

### 任务列表
- [x] 已完成的任务
- [ ] 未完成的任务
- [ ] 另一个待办事项

## 4. 代码块

### JavaScript 代码
```javascript
function fibonacci(n) {
  if (n <= 1) return n;
  return fibonacci(n - 1) + fibonacci(n - 2);
}

console.log(fibonacci(10)); // 输出: 55
```

### Python 代码
```python
def quick_sort(arr):
    if len(arr) <= 1:
        return arr
    pivot = arr[len(arr) // 2]
    left = [x for x in arr if x < pivot]
    middle = [x for x in arr if x == pivot]
    right = [x for x in arr if x > pivot]
    return quick_sort(left) + middle + quick_sort(right)

print(quick_sort([3, 6, 8, 10, 1, 2, 1]))
```

### Go 代码
```go
package main

import "fmt"

func main() {
    ch := make(chan int, 2)
    ch <- 1
    ch <- 2
    fmt.Println(<-ch)
    fmt.Println(<-ch)
}
```

## 5. 引用

> 这是一个引用块
>
> 可以包含多行内容
>
> > 这是嵌套引用
> >
> > 可以多层嵌套

## 6. 表格

| 姓名 | 年龄 | 职业 | 城市 |
|------|------|------|------|
| 张三 | 28 | 工程师 | 北京 |
| 李四 | 32 | 设计师 | 上海 |
| 王五 | 25 | 产品经理 | 深圳 |

### 对齐表格

| 左对齐 | 居中对齐 | 右对齐 |
|:-------|:--------:|-------:|
| 内容1 | 内容2 | 内容3 |
| 长内容 | 短 | 中等长度 |

## 7. 链接和图片

[这是一个链接](https://github.com)

[带标题的链接](https://google.com "Google 搜索")

自动链接: https://www.example.com

![图片描述](https://via.placeholder.com/150)

## 8. 数学公式

### 行内公式
这是一个行内公式 $E = mc^2$，爱因斯坦的质能方程。

圆的面积公式：$A = \pi r^2$

### 块级公式

$$
\int_0^\infty e^{-x^2} dx = \frac{\sqrt{\pi}}{2}
$$

$$
\sum_{i=1}^{n} i = \frac{n(n+1)}{2}
$$

$$
\begin{bmatrix}
a & b \\
c & d
\end{bmatrix}
\begin{bmatrix}
x \\
y
\end{bmatrix}
=
\begin{bmatrix}
ax + by \\
cx + dy
\end{bmatrix}
$$

## 9. 分隔线

---

***

___

## 10. HTML 支持

<div style="color: blue; font-weight: bold;">
这是一段蓝色粗体文本（使用 HTML）
</div>

<details>
<summary>点击展开详情</summary>

这是隐藏的内容，点击上方可以展开或折叠。

- 列表项 1
- 列表项 2
- 列表项 3

</details>

## 11. 脚注

这是一段包含脚注的文本[^1]。

这是另一个脚注[^note]。

[^1]: 这是第一个脚注的内容
[^note]: 这是命名脚注的内容

## 12. 定义列表

术语 1
: 定义 1

术语 2
: 定义 2a
: 定义 2b

## 13. 转义字符

使用反斜杠转义特殊字符：\* \_ \[ \] \( \) \# \+ \- \. \!

## 14. 复杂嵌套示例

1. **第一个主要步骤**

   这是第一步的详细说明。

   ```javascript
   const step1 = () => {
     console.log("执行第一步");
   };
   ```

   - 子任务 1.1
   - 子任务 1.2
     > 这是一个重要提示

2. **第二个主要步骤**

   | 参数 | 类型 | 说明 |
   |------|------|------|
   | name | string | 名称 |
   | age | number | 年龄 |

   公式：$f(x) = x^2 + 2x + 1$

3. **第三个主要步骤**

   - [x] 完成设计
   - [x] 完成开发
   - [ ] 完成测试

## 15. 特殊符号和 Emoji

支持 emoji: 😀 🎉 🚀 💡 ✅ ❌ ⚠️

特殊符号: © ® ™ § ¶ † ‡

箭头: → ← ↑ ↓ ⇒ ⇐

## 总结

这个文档包含了大部分常用的 Markdown 语法，可以用来测试渲染器的完整性。
