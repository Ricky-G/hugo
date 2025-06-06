// Copyright 2024 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transform_test

import (
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gohugoio/hugo/hugolib"
)

// Issue #11698
func TestMarkdownifyIssue11698(t *testing.T) {
	t.Parallel()

	files := `
-- config.toml --
disableKinds = ['home','section','rss','sitemap','taxonomy','term']
[markup.goldmark.parser.attribute]
title = true
block = true
-- layouts/_default/single.html --
_{{ markdownify .RawContent }}_
-- content/p1.md --
---
title: p1
---
foo bar
-- content/p2.md --
---
title: p2
---
foo

**bar**
-- content/p3.md --
---
title: p3
---
## foo

bar
-- content/p4.md --
---
title: p4
---
foo
{#bar}
  `

	b := hugolib.Test(t, files)

	b.AssertFileContent("public/p1/index.html", "_foo bar_")
	b.AssertFileContent("public/p2/index.html", "_<p>foo</p>\n<p><strong>bar</strong></p>\n_")
	b.AssertFileContent("public/p3/index.html", "_<h2 id=\"foo\">foo</h2>\n<p>bar</p>\n_")
	b.AssertFileContent("public/p4/index.html", "_<p id=\"bar\">foo</p>\n_")
}

func TestXMLEscape(t *testing.T) {
	t.Parallel()

	files := `
-- config.toml --
disableKinds = ['section','sitemap','taxonomy','term']
-- content/p1.md --
---
title: p1
---
a **b** ` + "\v" + ` c
<!--more-->
  `
	b := hugolib.Test(t, files)

	b.AssertFileContent("public/index.xml", `
	<description>&lt;p&gt;a &lt;strong&gt;b&lt;/strong&gt;  c&lt;/p&gt;</description>
	`)
}

// Issue #9642
func TestHighlightError(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
disableKinds = ['page','rss','section','sitemap','taxonomy','term']
-- layouts/index.html --
{{ highlight "a" "b" 0 }}
  `
	b := hugolib.NewIntegrationTestBuilder(
		hugolib.IntegrationTestConfig{
			T:           t,
			TxtarString: files,
		},
	)

	_, err := b.BuildE()
	b.Assert(err.Error(), qt.Contains, "error calling highlight: invalid Highlight option: 0")
}

// Issue #11884
func TestUnmarshalCSVLazyDecoding(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
disableKinds = ['page','rss','section','sitemap','taxonomy','term']
-- assets/pets.csv --
name,description,age
Spot,a nice dog,3
Rover,"a big dog",5
Felix,a "malicious" cat,7
Bella,"an "evil" cat",9
Scar,"a "dead cat",11
-- layouts/index.html --
{{ $opts := dict "lazyQuotes" true }}
{{ $data := resources.Get "pets.csv" | transform.Unmarshal $opts }}
{{ printf "%v" $data | safeHTML }}
  `
	b := hugolib.Test(t, files)

	b.AssertFileContent("public/index.html", `
[[name description age] [Spot a nice dog 3] [Rover a big dog 5] [Felix a "malicious" cat 7] [Bella an "evil" cat 9] [Scar a "dead cat 11]]
	`)
}

func TestToMath(t *testing.T) {
	files := `
-- hugo.toml --
disableKinds = ['page','rss','section','sitemap','taxonomy','term']
-- layouts/index.html --
{{ transform.ToMath "c = \\pm\\sqrt{a^2 + b^2}" }}
  `
	b := hugolib.Test(t, files)

	b.AssertFileContent("public/index.html", `
<span class="katex"><math
	`)
}

func TestToMathError(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		files := `
-- hugo.toml --
disableKinds = ['page','rss','section','sitemap','taxonomy','term']
-- layouts/index.html --
{{  transform.ToMath "c = \\foo{a^2 + b^2}" }}
  `
		b, err := hugolib.TestE(t, files, hugolib.TestOptWarn())

		b.Assert(err, qt.IsNotNil)
		b.Assert(err.Error(), qt.Contains, "KaTeX parse error: Undefined control sequence: \\foo")
	})

	t.Run("Disable ThrowOnError", func(t *testing.T) {
		files := `
-- hugo.toml --
disableKinds = ['page','rss','section','sitemap','taxonomy','term']
-- layouts/index.html --
{{ $opts := dict "throwOnError" false }}
{{  transform.ToMath "c = \\foo{a^2 + b^2}" $opts }}
  `
		b, err := hugolib.TestE(t, files, hugolib.TestOptWarn())

		b.Assert(err, qt.IsNil)
		b.AssertFileContent("public/index.html", `#cc0000`) // Error color
	})

	t.Run("Handle in template", func(t *testing.T) {
		files := `
-- hugo.toml --
disableKinds = ['page','rss','section','sitemap','taxonomy','term']
-- layouts/index.html --
{{ with try (transform.ToMath "c = \\foo{a^2 + b^2}") }}
	{{ with .Err }}
	 	{{ warnf "error: %s" . }}
	{{ else }}
		{{ .Value }}
	{{ end }}
{{ end }}
  `
		b, err := hugolib.TestE(t, files, hugolib.TestOptWarn())

		b.Assert(err, qt.IsNil)
		b.AssertLogContains("WARN  error: template: index.html:1:22: executing \"index.html\" at <transform.ToMath>: error calling ToMath: KaTeX parse error: Undefined control sequence: \\foo at position 5: c = \\̲f̲o̲o̲{a^2 + b^2}")
	})

	// See issue 13239.
	t.Run("Handle in template, old Err construct", func(t *testing.T) {
		files := `
-- hugo.toml --
disableKinds = ['page','rss','section','sitemap','taxonomy','term']
-- layouts/index.html --
{{ with transform.ToMath "c = \\pm\\sqrt{a^2 + b^2}" }}
	{{ with .Err }}
	 	{{ warnf "error: %s" . }}
	{{ else }}
		{{ . }}
	{{ end }}
{{ end }}
  `
		b, err := hugolib.TestE(t, files, hugolib.TestOptWarn())

		b.Assert(err, qt.IsNotNil)
		b.Assert(err.Error(), qt.Contains, "the return type of transform.ToMath was changed in Hugo v0.141.0 and the error handling replaced with a new try keyword, see https://gohugo.io/functions/go-template/try/")
	})
}

func TestToMathBigAndManyExpressions(t *testing.T) {
	filesTemplate := `
-- hugo.toml --
disableKinds = ['rss','section','sitemap','taxonomy','term']
[markup.goldmark.extensions.passthrough]
enable = true
[markup.goldmark.extensions.passthrough.delimiters]
block  = [['\[', '\]'], ['$$', '$$']]
inline = [['\(', '\)'], ['$', '$']]
-- content/p1.md --
P1_CONTENT
-- layouts/index.html --
Home.
-- layouts/_default/single.html --
Content: {{ .Content }}|
-- layouts/_default/_markup/render-passthrough.html --
{{ $opts := dict "throwOnError" false "displayMode" true }}
{{ transform.ToMath .Inner $opts }}
  `

	t.Run("Very large file with many complex KaTeX expressions", func(t *testing.T) {
		files := strings.ReplaceAll(filesTemplate, "P1_CONTENT", "sourcefilename: testdata/large-katex.md")
		b := hugolib.Test(t, files)
		b.AssertFileContent("public/p1/index.html", `
		<span class="katex"><math
			`)
	})

	t.Run("Large and complex expression", func(t *testing.T) {
		// This is pulled from the file above, which times out for some reason.
		largeAndComplexeExpressions := `\begin{align*} \frac{\pi^2}{6}&=\frac{4}{3}\frac{(\arcsin 1)^2}{2}\\ &=\frac{4}{3}\int_0^1\frac{\arcsin x}{\sqrt{1-x^2}}\,dx\\ &=\frac{4}{3}\int_0^1\frac{x+\sum_{n=1}^{\infty}\frac{(2n-1)!!}{(2n)!!}\frac{x^{2n+1}}{2n+1}}{\sqrt{1-x^2}}\,dx\\ &=\frac{4}{3}\int_0^1\frac{x}{\sqrt{1-x^2}}\,dx +\frac{4}{3}\sum_{n=1}^{\infty}\frac{(2n-1)!!}{(2n)!!(2n+1)}\int_0^1x^{2n}\frac{x}{\sqrt{1-x^2}}\,dx\\ &=\frac{4}{3}+\frac{4}{3}\sum_{n=1}^{\infty}\frac{(2n-1)!!}{(2n)!!(2n+1)}\left[\frac{(2n)!!}{(2n+1)!!}\right]\\ &=\frac{4}{3}\sum_{n=0}^{\infty}\frac{1}{(2n+1)^2}\\ &=\frac{4}{3}\left(\sum_{n=1}^{\infty}\frac{1}{n^2}-\frac{1}{4}\sum_{n=1}^{\infty}\frac{1}{n^2}\right)\\ &=\sum_{n=1}^{\infty}\frac{1}{n^2} \end{align*}`
		files := strings.ReplaceAll(filesTemplate, "P1_CONTENT", fmt.Sprintf(`---
title: p1
---

$$%s$$
	`, largeAndComplexeExpressions))

		b := hugolib.Test(t, files)
		b.AssertFileContent("public/p1/index.html", `
		<span class="katex"><math
			`)
	})
}

// Issue #13406.
func TestToMathRenderHookPosition(t *testing.T) {
	filesTemplate := `
-- hugo.toml --
disableKinds = ['rss','section','sitemap','taxonomy','term']
[markup.goldmark.extensions.passthrough]
enable = true
[markup.goldmark.extensions.passthrough.delimiters]
block  = [['\[', '\]'], ['$$', '$$']]
inline = [['\(', '\)'], ['$', '$']]
-- content/p1.md --
---
title: p1
---

Block:

$$1+2$$

Some inline $1+3$ math.

-- layouts/index.html --
Home.
-- layouts/_default/single.html --
Content: {{ .Content }}|
-- layouts/_default/_markup/render-passthrough.html --
{{ $opts := dict "throwOnError" true "displayMode" true }}
{{- with try (transform.ToMath .Inner $opts ) }}
  {{- with .Err }}
    {{ errorf "KaTeX: %s: see %s." . $.Position }}
  {{- else }}
    {{- .Value }}
  {{- end }}
{{- end -}}

`

	// Block math.
	files := strings.Replace(filesTemplate, "$$1+2$$", "$$\\foo1+2$$", 1)
	b, err := hugolib.TestE(t, files)
	b.Assert(err, qt.IsNotNil)
	b.AssertLogContains("p1.md:6:1")

	// Inline math.
	files = strings.Replace(filesTemplate, "$1+3$", "$\\foo1+3$", 1)
	b, err = hugolib.TestE(t, files)
	b.Assert(err, qt.IsNotNil)
	b.AssertLogContains("p1.md:8:13")
}

func TestToMathMacros(t *testing.T) {
	files := `
-- hugo.toml --
disableKinds = ['page','rss','section','sitemap','taxonomy','term']
-- layouts/index.html --
{{ $macros := dict
    "\\addBar" "\\bar{#1}"
	"\\bold" "\\mathbf{#1}"
}}
{{ $opts := dict "macros" $macros }}
{{ transform.ToMath "\\addBar{y} + \\bold{H}" $opts }}
  `
	b := hugolib.Test(t, files)

	b.AssertFileContent("public/index.html", `
<mi>y</mi>
	`)
}

// Issue #12977
func TestUnmarshalWithIndentedYAML(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
disableKinds = ['page','rss','section','sitemap','taxonomy','term']
-- layouts/index.html --
{{ $yaml := "\n  a:\n    b: 1\n  c:\n    d: 2\n" }}
{{ $yaml | transform.Unmarshal | encoding.Jsonify }}
`

	b := hugolib.Test(t, files)

	b.AssertFileExists("public/index.html", true)
	b.AssertFileContent("public/index.html", `{"a":{"b":1},"c":{"d":2}}`)
}

func TestPortableText(t *testing.T) {
	files := `
-- hugo.toml --
-- assets/sample.json --
[
  {
    "_key": "a",
    "_type": "block",
    "children": [
      {
        "_key": "b",
        "_type": "span",
        "marks": [],
        "text": "Heading 2"
      }
    ],
    "markDefs": [],
    "style": "h2"
  }
]
-- layouts/index.html --
{{ $markdown := resources.Get "sample.json" | transform.Unmarshal | transform.PortableText }}
Markdown: {{ $markdown }}|

`
	b := hugolib.Test(t, files)

	b.AssertFileContent("public/index.html", "Markdown: ## Heading 2\n|")
}

func TestUnmarshalCSV(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
disableKinds = ['page','rss','section','sitemap','taxonomy','term']
-- layouts/all.html --
{{ $opts := OPTS }}
{{ with resources.Get "pets.csv" | transform.Unmarshal $opts }}
  {{ jsonify . }}
{{ end }}
-- assets/pets.csv --
DATA
`

	// targetType = map
	f := strings.ReplaceAll(files, "OPTS", `dict "targetType" "map"`)
	f = strings.ReplaceAll(f, "DATA",
		"name,type,breed,age\nSpot,dog,Collie,3\nFelix,cat,Malicious,7",
	)
	b := hugolib.Test(t, f)
	b.AssertFileContent("public/index.html",
		`[{"age":"3","breed":"Collie","name":"Spot","type":"dog"},{"age":"7","breed":"Malicious","name":"Felix","type":"cat"}]`,
	)

	// targetType = map (no data)
	f = strings.ReplaceAll(files, "OPTS", `dict "targetType" "map"`)
	f = strings.ReplaceAll(f, "DATA", "")
	b = hugolib.Test(t, f)
	b.AssertFileContent("public/index.html", "")

	// targetType = slice
	f = strings.ReplaceAll(files, "OPTS", `dict "targetType" "slice"`)
	f = strings.ReplaceAll(f, "DATA",
		"name,type,breed,age\nSpot,dog,Collie,3\nFelix,cat,Malicious,7",
	)
	b = hugolib.Test(t, f)
	b.AssertFileContent("public/index.html",
		`[["name","type","breed","age"],["Spot","dog","Collie","3"],["Felix","cat","Malicious","7"]]`,
	)

	// targetType = slice (no data)
	f = strings.ReplaceAll(files, "OPTS", `dict "targetType" "slice"`)
	f = strings.ReplaceAll(f, "DATA", "")
	b = hugolib.Test(t, f)
	b.AssertFileContent("public/index.html", "")

	// targetType not specified
	f = strings.ReplaceAll(files, "OPTS", "dict")
	f = strings.ReplaceAll(f, "DATA",
		"name,type,breed,age\nSpot,dog,Collie,3\nFelix,cat,Malicious,7",
	)
	b = hugolib.Test(t, f)
	b.AssertFileContent("public/index.html",
		`[["name","type","breed","age"],["Spot","dog","Collie","3"],["Felix","cat","Malicious","7"]]`,
	)

	// targetType not specified (no data)
	f = strings.ReplaceAll(files, "OPTS", "dict")
	f = strings.ReplaceAll(f, "DATA", "")
	b = hugolib.Test(t, f)
	b.AssertFileContent("public/index.html", "")

	// targetType = foo
	f = strings.ReplaceAll(files, "OPTS", `dict "targetType" "foo"`)
	_, err := hugolib.TestE(t, f)
	if err == nil {
		t.Errorf("expected error")
	} else {
		if !strings.Contains(err.Error(), `invalid targetType: expected either slice or map, received foo`) {
			t.Log(err.Error())
			t.Errorf("error message does not match expected error message")
		}
	}

	// targetType = foo (no data)
	f = strings.ReplaceAll(files, "OPTS", `dict "targetType" "foo"`)
	f = strings.ReplaceAll(f, "DATA", "")
	_, err = hugolib.TestE(t, f)
	if err == nil {
		t.Errorf("expected error")
	} else {
		if !strings.Contains(err.Error(), `invalid targetType: expected either slice or map, received foo`) {
			t.Log(err.Error())
			t.Errorf("error message does not match expected error message")
		}
	}

	// targetType = map (error: expected at least a header row and one data row)
	f = strings.ReplaceAll(files, "OPTS", `dict "targetType" "map"`)
	_, err = hugolib.TestE(t, f)
	if err == nil {
		t.Errorf("expected error")
	} else {
		if !strings.Contains(err.Error(), `expected at least a header row and one data row`) {
			t.Log(err.Error())
			t.Errorf("error message does not match expected error message")
		}
	}

	// targetType = map (error: header row contains duplicate field names)
	f = strings.ReplaceAll(files, "OPTS", `dict "targetType" "map"`)
	f = strings.ReplaceAll(f, "DATA",
		"name,name,breed,age\nSpot,dog,Collie,3\nFelix,cat,Malicious,7",
	)
	_, err = hugolib.TestE(t, f)
	if err == nil {
		t.Errorf("expected error")
	} else {
		if !strings.Contains(err.Error(), `header row contains duplicate field names`) {
			t.Log(err.Error())
			t.Errorf("error message does not match expected error message")
		}
	}
}

// Issue 13729
func TestToMathStrictMode(t *testing.T) {
	t.Parallel()

	files := `
-- hugo.toml --
disableKinds = ['page','rss','section','sitemap','taxonomy','term']
-- layouts/all.html --
{{ transform.ToMath "a %" dict }}
-- foo --
`

	// strict mode: default
	f := strings.ReplaceAll(files, "dict", "")
	b, err := hugolib.TestE(t, f)
	b.Assert(err.Error(), qt.Contains, "[commentAtEnd]")

	// strict mode: error
	f = strings.ReplaceAll(files, "dict", `(dict "strict" "error")`)
	b, err = hugolib.TestE(t, f)
	b.Assert(err.Error(), qt.Contains, "[commentAtEnd]")

	// strict mode: ignore
	f = strings.ReplaceAll(files, "dict", `(dict "strict" "ignore")`)
	b = hugolib.Test(t, f, hugolib.TestOptWarn())
	b.AssertLogMatches("")
	b.AssertFileContent("public/index.html", `<annotation encoding="application/x-tex">a %</annotation>`)

	// strict: warn
	f = strings.ReplaceAll(files, "dict", `(dict "strict" "warn")`)
	b = hugolib.Test(t, f, hugolib.TestOptWarn())
	b.AssertLogMatches("[commentAtEnd]")
	b.AssertFileContent("public/index.html", `<annotation encoding="application/x-tex">a %</annotation>`)

	// strict mode: invalid value
	f = strings.ReplaceAll(files, "dict", `(dict "strict" "foo")`)
	b, err = hugolib.TestE(t, f)
	b.Assert(err.Error(), qt.Contains, "invalid strict mode")
}
