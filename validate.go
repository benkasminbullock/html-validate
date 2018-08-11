/* Validate HTML. */

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"unicode/utf8"
)

type lineTag struct {
	tag  string
	line int
}

/* Make a map of valid words. */

func makeMap(list []string) (z map[string]bool) {
	z = make(map[string]bool)
	for _, tag := range list {
		z[tag] = true
	}
	return z
}

/* Set up all the tables for validation. */

var valid map[string]bool
var noClose map[string]bool
var nonNesting map[string]bool
var nonEmpty map[string]bool

/* Initialise the tables of tags. */

func initTagTables() {
	valid = makeMap(validTags)
	noClose = makeMap(noCloseTags)
	nonNesting = makeMap(nonNestingTags)
	nonEmpty = makeMap(nonEmptyTags)
	for _, tag := range noCloseTags {
		valid[tag] = true
	}
	for _, tag := range nonNestingTags {
		valid[tag] = true
	}
}

/* A store of the current position within a string, used for
   navigating within HTML. */

type stringPos struct {
	position int
	i        int
	line     int
	filename string
	nletters int
}

func (sp *stringPos) String() string {
	return fmt.Sprintf("%s:%d:", sp.filename, sp.line)
}

func (sp *stringPos) Add(c string) {
	sp.i++
	sp.position += len(c)
}

// Jump over jumpbytes bytes, adjusting the strings and counting
// newlines.

func jumpover(html string, jumpbytes int, sp *stringPos) {
	jump := html[(*sp).position : (*sp).position+jumpbytes]
	addlines := strings.Count(jump, "\n")
	sp.line += addlines
	sp.i += utf8.RuneCountInString(jump)
	sp.position += jumpbytes
}

/* Find instances of tag IDs. */

func findIds(ids map[string]lineTag, tag string, remaining string, sp *stringPos) {
	// The located ID.
	var id string
	// This should allow for spaces, shouldn't it?
	idequal := strings.Index(remaining, "id=")
	if idequal != -1 {
		openquote := strings.IndexAny(remaining[idequal+1:], "\"'")
		if openquote != -1 {
			closequote := strings.IndexAny(remaining[idequal+openquote+2:], "\"'")
			if closequote != -1 {
				id = remaining[idequal+openquote+2 : idequal+openquote+2+closequote]
				if lt, exists := ids[id]; exists {
					fmt.Printf("%s duplicate id '%s'\n",
						sp.String(), id)
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

func findTag(html string, sp *stringPos, open bool, ids map[string]lineTag) (tag string, err error) {
	start := strings.IndexAny(html[sp.position:], " >\n")
	if start == -1 {
		return "", errors.New("No ending marker found")
	}
	tag = html[(*sp).position : (*sp).position+start]
	end := strings.IndexAny(html[(*sp).position:], ">")
	if end == -1 {
		return "", errors.New("No > after <")
	}

	// ids as nil is used for end tags.
	if open && start+1 < end {
		remaining := html[sp.position+start+1 : sp.position+end]
		if len(remaining) > 0 {
			findIds(ids, tag, remaining, sp)
		}
	}
	jumpover(html, end, sp)
	return tag, err
}

/* Skip over HTML comments or doctype declarations. */

func skipComment(html string, sp *stringPos) (err error) {
	var end int
	start := html[sp.position:]
	if strings.ToLower(start[0:7]) == "doctype" {
		end = strings.Index(start, ">")
		if end == -1 {
			return errors.New("No > found for doctype")
		}
	} else if start[0:2] == "--" {
		end = strings.Index(start, "-->")
		if end == -1 {
			return errors.New("No --> found for comment")
		}
	} else {
		return errors.New("<! but no doctype or comment marker")
	}
	jumpover(html, end, sp)
	return
}

/* Skip over <script>..</script>. */

func skipScript(html string, sp *stringPos) (err error) {
	var end int
	start := html[sp.position:]
	end = strings.Index(start, "</script>")
	if end == -1 {
		fmt.Printf("%s no </script> found.\n", sp)
		return errors.New("No </script> found for <script>")
	}
	// This leaves the </script> tag at the start of what sp points
	// to, then it is handled by the usual closing tag stuff.
	jumpover(html, end, sp)
	return
}

/* Validate the HTML in "html", which comes from a file with the name
   "filename". */

func validate(html string, filename string) {
	var opentags []lineTag
	nestTags := make(map[string]lineTag)
	// id= things within HTML opening tags.
	ids := make(map[string]lineTag)
	var sp stringPos
	split := strings.Split(html, "")
	sp.nletters = len(split)
	// Number of the letter.
	sp.i = 0
	// The current line number.
	sp.line = 1
	sp.filename = filename
	for sp.i < sp.nletters {
		c := split[sp.i]
		switch c {
		case "<":
			sp.Add(c)
			c := split[sp.i]
			switch c {
			case "/":

				/* Closing tag, pop the stack "opentags" to find a
				/* match. */

				sp.Add(c)

				tag, err := findTag(html, &sp, false, ids)
				if err != nil {
					fmt.Printf("%s error %s\n",
						sp.String(), err)
					break
				}
				delete(nestTags, tag)
				var toptag lineTag
				if len(opentags) > 0 {
					opentags, toptag = opentags[:len(opentags)-1], opentags[len(opentags)-1]
				} else {
					fmt.Printf("%s too many closing tags.\n",
						sp.String())
				}
				if toptag.tag != tag {
					closed := false
					for scrape := len(opentags) - 1; scrape >= 0; scrape-- {
						scrapeTag := opentags[scrape]
						if scrapeTag.tag == tag {
							fmt.Printf("%s tag mismatch: <%s> </%s>: ",
								sp.String(), toptag.tag, tag)
							fmt.Printf("popping %d unclosed tags.\n",
								len(opentags)-scrape)
							for i := scrape + 1; i < len(opentags); i++ {
								fmt.Printf("%s:%d: <%s> unclosed.\n",
									sp.filename, opentags[i].line, opentags[i].tag)
							}
							fmt.Printf("%s:%d: <%s> unclosed.\n",
								sp.filename, toptag.line, toptag.tag)
							opentags = opentags[:scrape]
							closed = true
							break
						}
					}
					if !closed {
						fmt.Printf("%s closing tag </%s> with no opening tag.\n",
							sp.String(), tag)
						// Push the last thing back on there.
						opentags = append(opentags, toptag)
					}
				}

			case " ":
				fmt.Printf("%s space character after <.\n",
					sp.String())
			case "!":
				sp.Add(c)
				err := skipComment(html, &sp)
				if err != nil {
					fmt.Printf("%s error %s\n",
						sp.String(), err)
				}
			default:
				tag, err := findTag(html, &sp, true, ids)
				if !valid[tag] {
					fmt.Printf("%s unknown tag <%s>.\n",
						sp.String(), tag)
				}
				if err != nil {
					fmt.Printf("%s error %s\n",
						sp.String(), err)
				} else {
					if !noClose[tag] {
						var lt lineTag
						lt.tag = tag
						lt.line = sp.line
						opentags = append(opentags, lt)
						if nonNesting[tag] {
							previous, nested := nestTags[tag]
							if nested {
								fmt.Printf("%s nested <%s>.\n",
									sp.String(), tag)
								fmt.Printf("%s:%d: first <%s> is here.\n",
									sp.filename, previous.line, tag)
							} else {
								nestTags[tag] = lt
							}
						}
					}
				}
				if tag == "script" {
					skipScript(html, &sp)
				}
			}
		case "\n":
			sp.line++
			fallthrough
		default:
			sp.Add(c)
		}
	}
	if len(opentags) > 0 {
		fmt.Printf("There are %d unclosed tags:\n", len(opentags))
		for i := 0; i < len(opentags); i++ {
			fmt.Printf("%s:%d: <%s>\n", sp.filename, opentags[i].line, opentags[i].tag)
		}
	}
}

func main() {
	initTagTables()
	for i := 1; i < len(os.Args); i++ {
		//http://stackoverflow.com/questions/13514184/how-can-i-read-a-whole-file-into-a-string-variable-in-golang#13514395
		file := os.Args[i]
		buf, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Println("error file ", err)
			continue
		}
		if !utf8.Valid(buf) {
			// This should be printing to stderr, not stdout.
			fmt.Fprintf(os.Stderr, "%s is not UTF-8.\n", file)
			continue
		}
		s := string(buf)
		validate(s, file)
	}
}
