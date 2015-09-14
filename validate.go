package main;

import (
	"fmt"
	"strings"
	"errors"
	"unicode/utf8"
	"io/ioutil"
	"os"
)

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
	"figcaption",
	"figure",
	"font",
	"form",
	"ins",
	"label",
	"li",
	"noscript",
	"ol",
	"option",
	"ref", // temp remove this just for sljfaq.ff checking.
	"rp",
	"rt",
	"ruby",
	"select",
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
	"dd",
	"dt",
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
	"p",
	"pre",
	"script",
	"small",
	"span",
	"span",
	"strong",
	"strong",
	"style",
	"title",
}

var valid map[string]bool
var noClose map[string]bool
var nonNesting map[string]bool

type lineTag struct {
	tag string
	line int
}

func makeMap(list []string) (z map[string]bool) {
	z = make(map [string]bool)
	for _, tag := range list {
		z[tag] = true
	}
	return z
}

/* Set up all the tables for validation. */

func initValid() {
	valid = makeMap(validTags)
	noClose = makeMap(noCloseTags)
	nonNesting = makeMap(nonNestingTags)
	for _, tag := range noCloseTags {
		valid[tag] = true
	}
	for _, tag := range nonNestingTags {
		valid[tag] = true
	}
}

type stringPos struct {
	position int
	i int
	line int
	filename string
}

func (sp *stringPos) String() string {
	return fmt.Sprintf("%s:%d:", sp.filename, sp.line)
}

// Jump over jumpbytes bytes, adjusting the strings and counting
// newlines.

func jumpover(html string, jumpbytes int, sp * stringPos) {
	jump := html[(*sp).position:(*sp).position+jumpbytes]
	addlines := strings.Count(jump, "\n")
/*
	fmt.Printf("Jumping over '%s'.\n", jump)
	fmt.Printf("Adding %d lines.\n", addlines)
*/
	sp.line += addlines
	sp.i += utf8.RuneCountInString(jump)
	sp.position += jumpbytes
}

func findIds(ids map[string]lineTag, tag string, remaining string, sp * stringPos) {
	// The located ID.
	var id string
	idequal := strings.Index(remaining, "id=")
	if idequal != -1 {
		openquote := strings.IndexAny(remaining[idequal+1:], "\"'")
		if openquote != -1 {
			closequote := strings.IndexAny(remaining[idequal+openquote+2:], "\"'")
			if closequote != -1 {
				id = remaining[idequal+openquote+2:idequal+openquote+2+closequote]
//				fmt.Printf("%s found id '%s'\n", sp, id);
				if lt, exists := ids[id]; exists {
					fmt.Printf("%s:%d: duplicate id '%s'\n",
						sp.filename, sp.line, id);
					fmt.Printf("%s:%d: previous instance in tag '%s'\n",
						sp.filename, lt.line, lt.tag) 
				} else {
					var lt lineTag
					lt.line = sp.line
					lt.tag = tag
					ids[id] = lt
				}
			}
		}
	}
}

func findTag(html string, sp * stringPos, open bool, ids map[string]lineTag) (tag string, err error) {
	start := strings.IndexAny(html[sp.position:], " >\n")
	if start == -1 {
		return "", errors.New("No ending marker found")
	}
//	fmt.Printf("%s %d %d\n",html[*position:], *position, start)
	tag = html[(*sp).position:(*sp).position+start]
	end := strings.IndexAny(html[(*sp).position:], ">")
	if end == -1 {
		return "", errors.New("No > after <")
	}

	// ids as nil is used for end tags.
	if open && start+1 < end {
		remaining := html[sp.position+start+1:sp.position+end]
		if len(remaining) > 0 {
			findIds(ids, tag, remaining, sp)
		}
	}
	jumpover(html, end, sp)
 	return tag, err
}

/* Skip over HTML comments or doctype declarations. */

func skipComment(html string, sp * stringPos) (err error) {
	var end int
	start := html[sp.position:]
	if strings.ToLower(start[0:7]) == "doctype" {
		end = strings.Index(start, ">")
		if (end == -1) {
			return errors.New("No > found for doctype")
		}
	} else if start[0:2] == "--" {
		end = strings.Index(start, "-->")
		if (end == -1) {
			return errors.New("No --> found for comment")
		}
	} else {
		return errors.New("<! but no doctype or comment marker")
	}
	jumpover (html, end, sp)
	return
}

/* Skip over <script>..</script>. */

func skipScript(html string, sp * stringPos) (err error) {
	var end int
	start := html[sp.position:]
	end = strings.Index(start, "</script>")
	if (end == -1) {
		fmt.Printf("%s no </script> found.\n", sp);
		return errors.New("No </script> found for <script>")
	}
	// This leaves the </script> tag at the start of what sp points
	// to, then it is handled by the usual closing tag stuff.
	jumpover (html, end, sp)
	return
}

func validate(html string, filename string, offset int) {
	var opentags []lineTag
	nestTags := make(map[string]lineTag)
	// id= things within HTML opening tags.
	ids := make(map[string]lineTag)
	var sp stringPos
	split := strings.Split(html, "")
	nletters := len(split)
	// Number of the letter.
	sp.i = 0
	// The current line number.
	sp.line = 1 + offset
	sp.filename = filename
	for sp.i < nletters {
		c := split[sp.i]
		switch c {
		case "<":
//			fmt.Printf("%s:%d: tag.\n", filename, line)
			sp.i++
			sp.position += len(c)
			c := split[sp.i]
			switch c {
			case "/":

				/* Closing tag, pop the stack "opentags" to find a
				/* match. */

				sp.i++
				sp.position += len(c)
				tag, err := findTag(html, & sp, false, ids)
				if err != nil {
					fmt.Printf("%s:%d: error %s\n",
						filename, sp.line, err)
				} else {
/*
					fmt.Printf("%s:%d: %s close tag %s\n",
						filename, sp.line, c, tag)
*/
					delete (nestTags, tag)
					var toptag lineTag
					if len(opentags) > 0 {
						opentags, toptag = opentags[:len(opentags) - 1], opentags[len(opentags) - 1]
					} else {
						fmt.Printf("%s:%d: too many closing tags.\n",
							filename, sp.line);
					}
					if (toptag.tag != tag) {
						closed := false
						for scrape := len(opentags) - 1; scrape >= 0; scrape-- {
							scrapeTag := opentags[scrape]
							if scrapeTag.tag == tag {
								fmt.Printf("%s:%d: tag mismatch: <%s> </%s>: ",
									filename, sp.line, toptag.tag, tag)
								fmt.Printf("popping %d unclosed tags.\n",
									len(opentags) - scrape)
								for i := scrape + 1; i < len(opentags); i++ {
									fmt.Printf("%s:%d: <%s> unclosed.\n",
										filename, opentags[i].line, opentags[i].tag);
								}
								fmt.Printf("%s:%d: <%s> unclosed.\n",
									filename, toptag.line, toptag.tag);
								opentags = opentags[:scrape]
								closed = true
								break
							}
						}
						if ! closed {
							fmt.Printf("%s:%d: closing tag </%s> with no opening tag.\n",
								filename, sp.line, tag)
							// Push the last thing back on there.
							opentags = append(opentags, toptag)
						}
					}

				}
			case " ":
				fmt.Printf("%s:%d: space character after <.\n",
					filename, sp.line)
			case "!":
				sp.i++
				sp.position += len(c)
				err := skipComment(html, & sp)
				if err != nil {
					fmt.Printf("%s:%d: error %s\n",
						filename, sp.line, err)
				}
			default:
				tag, err := findTag(html, & sp, true, ids)
				if ! valid[tag] {
					fmt.Printf("%s:%d: unknown tag <%s>.\n",
						filename, sp.line, tag)
				}
				if err != nil {
					fmt.Printf("%s:%d: error %s\n",
						filename, sp.line, err)
				} else {
//					fmt.Printf("%s:%d: open tag %s.\n", filename, sp.line, tag)
					if ! noClose[tag] {
						var lt lineTag
						lt.tag = tag
						lt.line = sp.line
						opentags = append(opentags, lt)
						if nonNesting[tag] {
							/*
							fmt.Printf("%s:%d: non-nesting tag %s\n",
								filename, sp.line, tag);
*/
							previous, nested := nestTags[tag]
							if nested {
								fmt.Printf("%s:%d: nested <%s>.\n",
									filename, sp.line, tag);
								fmt.Printf("%s:%d: first <%s> is here.\n",
									filename, previous.line, tag)
							} else {
								nestTags[tag] = lt;
							}
						}
					}
				}
				if tag == "script" {
					skipScript(html, & sp)
				}
			}
		case "\n":
			sp.line++;
			fallthrough
		default:
			sp.i++
			sp.position += len(c)
		}
	}
	if len(opentags) > 0 {
		fmt.Printf("There are %d unclosed tags:\n", len(opentags))
		for i := 0; i < len(opentags); i++ {
			fmt.Printf("%s:%d: <%s>\n", filename, opentags[i].line, opentags[i].tag)
		}
	}
}
//http://stackoverflow.com/questions/13514184/how-can-i-read-a-whole-file-into-a-string-variable-in-golang#13514395

func main() {
	initValid()
	for i := 1; i < len(os.Args); i++ {
		file := os.Args[i]
		buf, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Println("error file ",err)
		} else {
			s := string(buf)
			validate (s, file, 0)
		}
	}
}

