// This is a hand-edited list. make-tag-go.pl does not work yet.

package main

// Valid HTML tags

var validTags = []string{
	"article",
	"aside",
	"big",
	"blockquote",
	"body",
	"button",
	"canvas",
	"caption",
	"dd",
	"del",
	"details",
	"div",
	"dl",
	"dt",
	"figcaption",
	"figure",
	"font",
	"footer",
	"form",
	"header",
	"ins",
	"kbd",
	"label",
	"li",
	"main",
	"mark",
	"nav",
	"nobr",
	"noscript",
	"ol",
	"option",
	"rp",
	"rt",
	"ruby",
	"section",
	"select",
	"span",
	"sub",
	"summary",
	"sup",
	"table",
	"tbody",
	"td",
	"textarea",
	"th",
	"time",
	"tr",
	"u",
	"ul",
}

/* These tags do not need an ending tag like </p> */

var noCloseTags = []string{
	"area",
	"br",
	"hr",
	"image",
	"input",
	"img",
	"link",
	"meta",
}

// These tags should not be nested, e.g. <p><p>. 

var nonNestingTags = []string {
	"a",
	"b",
	"center",
	"code",
	"em",
	"h1",
	"h2",
	"h3",
	"h4",
	"head",
	"html",
	"i",
	"map",
	"nav",
	"p",
	"pre",
	"script",
	"small",
	"strong",
	"style",
	"title",
}

var nonEmptyTags = []string {
"p",
//"a",
}
