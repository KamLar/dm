package dm_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/assertly"
	"github.com/viant/dm"
	"github.com/viant/toolbox"
	"os"
	"path"
	"testing"
)

func TestDOM(t *testing.T) {
	testLocation := toolbox.CallerDirectory(3)

	type attrSearch struct {
		attribute string
		oldValue  string
		newValue  string
		fullMatch bool
		tag       string
	}

	type innerHTMLSearch struct {
		attribute string
		tag       string
		value     string
	}

	testcases := []struct {
		description   string
		uri           string
		attributes    *dm.Filters
		newAttributes []attrSearch
		innerHTMLGet  []innerHTMLSearch
		innerHTMLSet  []innerHTMLSearch
	}{
		{
			uri: "template001",
		},
		{
			uri: "template002",
			newAttributes: []attrSearch{
				{tag: "img", attribute: "src", oldValue: "[src]", newValue: "abcdef", fullMatch: true},
			},
		},
		{
			uri: "template003",
			newAttributes: []attrSearch{
				{tag: "img", attribute: "src", oldValue: "[src]", newValue: "abcdef", fullMatch: true},
				{tag: "p", attribute: "class", oldValue: "[class]", newValue: "newClasses", fullMatch: true},
				{tag: "div", attribute: "hidden", oldValue: "[hidden]", newValue: "newHidden", fullMatch: true},
			},
			innerHTMLGet: []innerHTMLSearch{
				{tag: "div", attribute: "hidden", value: `This is div inner`},
			},
			innerHTMLSet: []innerHTMLSearch{
				{tag: "div", attribute: "hidden", value: `<iframe hidden><div class="div-class">Hello world</div></iframe>`},
			},
		},
		{
			uri: "template004",
			newAttributes: []attrSearch{
				{tag: "img", attribute: "src", oldValue: "[src]", newValue: "newSrc", fullMatch: true},
				{tag: "img", attribute: "src", oldValue: "newSrc", newValue: "abcdef", fullMatch: true},
			},
			innerHTMLGet: []innerHTMLSearch{
				{tag: "head", value: `
    <title>Index</title>
`},
			},
		},
		{
			uri: "template005",
			newAttributes: []attrSearch{
				{tag: "img", attribute: "src", oldValue: "[src]", newValue: "newSrc"},
				{tag: "img", attribute: "alt", oldValue: "alt", newValue: "newAlt"},
			},
			innerHTMLGet: []innerHTMLSearch{
				{tag: "head", value: ``},
			},

			attributes: dm.NewFilters(
				dm.NewFilter("img", "src"),
			),
		},
		{
			uri: "template008",
			innerHTMLGet: []innerHTMLSearch{
				{tag: "script", value: ``},
			},

			attributes: dm.NewFilters(
				dm.NewFilter("script"),
			),
		},
	}

	//for _, testcase := range testcases[len(testcases)-1:] {
	for _, testcase := range testcases {
		templatePath := path.Join(testLocation, "testdata", testcase.uri)
		dom, err := readFromFile(path.Join(templatePath, "index.html"), testcase.attributes)
		if !assert.Nil(t, err, testcase.description) {
			t.Fail()
			continue
		}

		session := dom.DOM()
		for _, newAttr := range testcase.newAttributes {
			attrIterator := session.SelectAttributes(newAttr.tag, newAttr.attribute, newAttr.oldValue)
			for attrIterator.Has() {
				attr, _ := attrIterator.Next()
				attr.Set(newAttr.newValue)
				assert.Equal(t, newAttr.newValue, attr.Value(), testcase.uri)
			}
		}

		for _, search := range testcase.innerHTMLGet {
			selectors := make([]string, 0)
			if search.tag != "" {
				selectors = append(selectors, search.tag)
			}

			if search.attribute != "" {
				selectors = append(selectors, search.attribute)
			}

			tagIt := session.Select(selectors...)
			for tagIt.Has() {
				element, _ := tagIt.Next()
				assertly.AssertValues(t, search.value, element.InnerHTML())
			}
		}

		for _, search := range testcase.innerHTMLSet {
			selectors := make([]string, 0)
			if search.tag != "" {
				selectors = append(selectors, search.tag)
			}

			if search.attribute != "" {
				selectors = append(selectors, search.attribute)
			}

			selectorIt := session.Select(selectors...)
			for selectorIt.Has() {
				element, _ := selectorIt.Next()
				_ = element.SetInnerHTML(search.value)
			}
		}

		result, err := os.ReadFile(path.Join(templatePath, "expect.html"))
		if !assert.Nil(t, err, testcase.uri) {
			t.Fail()
			continue
		}
		bytes := session.Render()
		assertly.AssertValues(t, bytes, result, testcase.uri)
	}
}

func TestDOM_Element_NewAttribute(t *testing.T) {
	templatePath := path.Join(toolbox.CallerDirectory(3), "testdata", "template007")
	vdom, err := readFromFile(path.Join(templatePath, "index.html"))
	if !assert.Nil(t, err) {
		return
	}

	dom := vdom.DOM()
	tagIt := dom.Select("p", "class", "p-100")
	for tagIt.Has() {
		aTag, _ := tagIt.Next()
		aTag.AddAttribute("id", "first-paragraph")
	}

	attrIt := dom.SelectAttributes("img", "src")
	for attrIt.Has() {
		attribute, _ := attrIt.Next()
		attribute.Set("some-img.jpg")
	}

	tagIt = dom.Select("div")
	for tagIt.Has() {
		aTag, _ := tagIt.Next()
		_ = aTag.SetInnerHTML("New inner html")
	}

	tagIt = dom.Select("p", "class", "p-small")
	for tagIt.Has() {
		aTag, _ := tagIt.Next()
		aTag.AddAttribute("id", "second-paragraph")
	}

	render := dom.Render()
	result, err := os.ReadFile(path.Join(templatePath, "expect.html"))
	if !assert.Nil(t, err) {
		return
	}
	assertly.AssertValues(t, string(result), render)
}

func readFromFile(path string, options ...dm.Option) (*dm.VirtualDOM, error) {
	template, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	dom, err := dm.New(string(template), options...)
	if err != nil {
		return nil, err
	}
	return dom, nil
}

//Benchmarks
var vdom *dm.VirtualDOM

func init() {
	var err error
	vdom, err = dm.New(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Title</title>
</head>
<body>
<img src="abc.jpg" alt="some img"/>
<img src="abc.jpg" alt="some img"/>
<img src="abc.jpg" alt="some img"/>
<iframe></iframe>
</body>
</html>`, dm.NewFilters(
		dm.NewFilter("img", "src"),
		dm.NewFilter("iframe"),
	))

	if err != nil {
		panic(err)
	}
}
func BenchmarkVirtualDOM_DOM(b *testing.B) {
	var result string
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dom := vdom.DOM()
		imgSrcIt := dom.SelectAttributes("img", "src")
		for imgSrcIt.Has() {
			attribute, _ := imgSrcIt.Next()
			attribute.Set("newSrc")
		}

		iframeIt := dom.Select("iframe")
		for iframeIt.Has() {
			iframe, _ := iframeIt.Next()
			_ = iframe.SetInnerHTML("this is new inner")
		}
		result = dom.Render()
	}

	assert.Equal(b, "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <title>Title</title>\n</head>\n<body>\n<img src=\"newSrc\" alt=\"some img\"/>\n<img src=\"newSrc\" alt=\"some img\"/>\n<img src=\"newSrc\" alt=\"some img\"/>\n<iframe>this is new inner</iframe>\n</body>\n</html>", result)
}
