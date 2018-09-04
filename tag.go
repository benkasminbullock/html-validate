// This is a hand-edited list. make-tag-go.pl does not work yet.

package main

// Valid HTML tags

var validTags = []string{
	"big",
	"blockquote",
	"body",
	"button",
	"canvas",
	"caption",
	"del",
	"div",
	"dl",
	"dd",
	"dt",
	"figcaption",
	"figure",
	"font",
	"form",
	"ins",
	"kbd",
	"label",
	"li",
	"noscript",
	"nobr",
	"ol",
	"option",
	"rp",
	"rt",
	"ruby",
	"select",
	"sub",
	"sup",
	"table",
	"tbody",
	"td",
	"textarea",
	"th",
	"tr",
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
	"span",
	"strong",
	"style",
	"title",
}

var nonEmptyTags = []string {
"p",
//"a",
}
