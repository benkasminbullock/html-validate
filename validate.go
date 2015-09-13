package main;

import (
	"fmt"
	"strings"
	"errors"
	//"runtime"
	"unicode/utf8"
	"io/ioutil"
	"os"
)

// Valid HTML tags

var validTags = []string{
	"a",
	"b",
	"blockquote",
	"body",
	"caption",
	"center",
	"code",
	"del",
	"div",
	"dl",
	"em",
	"figcaption",
	"figure",
	"font",
	"form",
	"h1",
	"h2",
	"h3",
	"h4",
	"head",
	"html",
	"i",
	"ins",
	"li",
	"map",
	"ol",
	"p",
	"pre",
	"ref", // temp remove this just for sljfaq.ff checking.
	"rt",
	"rp",
	"ruby",
	"script",
	"small",
	"span",
	"strong",
	"sup",
	"table",
	"tbody",
	"td",
	"th",
	"title",
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

var valid map[string]bool
var noClose map[string]bool

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

func initValid() {
	valid = makeMap(validTags)
	noClose = makeMap(noCloseTags)
	for _, tag := range noCloseTags {
		valid[tag] = true
	}
}

type stringPos struct {
	position int
	i int
	line int
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

func findTag(html string, sp * stringPos) (tag string, err error) {
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
	jumpover(html, end, sp)
 	return tag, err
}

func skipComment(html string, sp * stringPos) (err error) {
	end := strings.Index(html[sp.position:], "-->")
	if (end == -1) {
		return errors.New("No --> found for comment")
	}
	jumpover (html, end, sp)
	return
}

func validate(html string, filename string, offset int) {
	var opentags []lineTag
	var sp stringPos
	split := strings.Split(html, "")
	nletters := len(split)
	// Number of the letter.
	sp.i = 0
	// The current line number.
	sp.line = 1 + offset
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
				tag, err := findTag(html, & sp)
				if err != nil {
					fmt.Printf("%s:%d: error %s\n",
						filename, sp.line, err)
				} else {
/*
					fmt.Printf("%s:%d: %s close tag %s\n",
						filename, sp.line, c, tag)
*/
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
				err := skipComment(html, & sp)
				if err != nil {
					fmt.Printf("%s:%d: error %s\n",
						filename, sp.line, err)
				}
			default:
				tag, err := findTag(html, & sp)
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
					}
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

