package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConvert(t *testing.T) {
	t.Parallel()

	type args struct {
		html string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// 見出し
		{
			name: "h1タグの場合に#に変換される",
			args: args{html: "<h1>Title</h1>"},
			want: "# Title",
		},
		{
			name: "h2タグの場合に##に変換される",
			args: args{html: "<h2>Subtitle</h2>"},
			want: "## Subtitle",
		},
		{
			name: "h3タグの場合に###に変換される",
			args: args{html: "<h3>Section</h3>"},
			want: "### Section",
		},
		{
			name: "h4タグの場合に####に変換される",
			args: args{html: "<h4>Subsection</h4>"},
			want: "#### Subsection",
		},
		{
			name: "h5タグの場合に#####に変換される",
			args: args{html: "<h5>Minor</h5>"},
			want: "##### Minor",
		},
		{
			name: "h6タグの場合に######に変換される",
			args: args{html: "<h6>Smallest</h6>"},
			want: "###### Smallest",
		},
		// 段落
		{
			name: "pタグの場合にテキストのみに変換される",
			args: args{html: "<p>Hello world</p>"},
			want: "Hello world",
		},
		{
			name: "複数のpタグの場合に空行で区切られる",
			args: args{html: "<p>First</p><p>Second</p>"},
			want: "First\n\nSecond",
		},
		// 強調
		{
			name: "strongタグの場合に**で囲まれる",
			args: args{html: "<strong>bold</strong>"},
			want: "**bold**",
		},
		{
			name: "bタグの場合に**で囲まれる",
			args: args{html: "<b>bold</b>"},
			want: "**bold**",
		},
		{
			name: "emタグの場合に*で囲まれる",
			args: args{html: "<em>italic</em>"},
			want: "*italic*",
		},
		{
			name: "iタグの場合に*で囲まれる",
			args: args{html: "<i>italic</i>"},
			want: "*italic*",
		},
		// リンク
		{
			name: "aタグの場合にMarkdownリンクに変換される",
			args: args{html: `<a href="https://example.com">Example</a>`},
			want: "[Example](https://example.com)",
		},
		// 画像
		{
			name: "imgタグでalt属性がある場合に画像記法に変換される",
			args: args{html: `<img src="image.png" alt="An image">`},
			want: "![An image](image.png)",
		},
		{
			name: "imgタグでalt属性がない場合に空のalt記法に変換される",
			args: args{html: `<img src="image.png">`},
			want: "![](image.png)",
		},
		// コード
		{
			name: "codeタグの場合にバッククォートで囲まれる",
			args: args{html: "<code>fmt.Println()</code>"},
			want: "`fmt.Println()`",
		},
		{
			name: "codeタグでHTMLエンティティがある場合にデコードされる",
			args: args{html: "<code>&lt;div&gt;</code>"},
			want: "`<div>`",
		},
		{
			name: "preとcodeタグの場合にコードブロックに変換される",
			args: args{html: "<pre><code>func main() {}</code></pre>"},
			want: "```\nfunc main() {}\n```",
		},
		// リスト
		{
			name: "ulとliタグの場合に箇条書きに変換される",
			args: args{html: "<ul><li>Item 1</li><li>Item 2</li></ul>"},
			want: "- Item 1\n- Item 2",
		},
		{
			name: "olとliタグの場合に番号付きリストに変換される",
			args: args{html: "<ol><li>First</li><li>Second</li></ol>"},
			want: "1. First\n2. Second",
		},
		// 引用
		{
			name: "blockquoteタグの場合に引用記法に変換される",
			args: args{html: "<blockquote>This is a quote</blockquote>"},
			want: "> This is a quote",
		},
		// テーブル
		{
			name: "tableタグの場合にMarkdownテーブルに変換される",
			args: args{html: `<table>
		<tr><th>Header 1</th><th>Header 2</th></tr>
		<tr><td>Cell 1</td><td>Cell 2</td></tr>
	</table>`},
			want: `| Header 1 | Header 2 |
| --- | --- |
| Cell 1 | Cell 2 |`,
		},
		// 水平線
		{
			name: "hrタグの場合に---に変換される",
			args: args{html: "<hr>"},
			want: "---",
		},
		{
			name: "自己終了hrタグの場合に---に変換される",
			args: args{html: "<hr/>"},
			want: "---",
		},
		// 改行
		{
			name: "brタグの場合に末尾2スペース改行に変換される",
			args: args{html: "Line 1<br>Line 2"},
			want: "Line 1  \nLine 2",
		},
		// 複合
		{
			name: "複数要素が混在する場合に正しく変換される",
			args: args{html: `<h1>Title</h1><p>This is <strong>bold</strong> and <em>italic</em>.</p>`},
			want: `# Title

This is **bold** and *italic*.`,
		},
		// 境界値
		{
			name: "空文字の場合に空文字を返す",
			args: args{html: ""},
			want: "",
		},
		{
			name: "HTMLタグがない場合にテキストのみを返す",
			args: args{html: "plain text"},
			want: "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Convert(tt.args.html)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Convert() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
