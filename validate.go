/* Validate HTML. */

package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

type lineTag struct {
	// The tag's name, like "html" for the <html> tag.
	tag string
	// The line number of the tag, so we can give good error messages.
	line int
	// The position of the tag within the string.
	position int
}

// Make a map of valid words.
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

// Initialise the tables of tags.
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
	// Position we are currently examining within the string
	position int
	// The number of the letter we are currently examining.
	i int
	// Line number within the string
	line int
	// File associated with the string
	filename string
	// The total number of letters in the string
	nletters int
}

// Show the file name and line number of the current position of "sp".
func (sp *stringPos) String() string {
	return fmt.Sprintf("%s:%d:", sp.filename, sp.line)
}

// Add the length of the string "c" to the position field of "sp" and
// increment the number of characters sp.i.
func (sp *stringPos) Add(c string) {
	sp.i++
	sp.position += len(c)
}

// Jump over "jumpbytes" bytes, adjusting the strings and counting
// newlines.
func jumpover(html string, jumpbytes int, sp *stringPos) {
	jump := html[(*sp).position : (*sp).position+jumpbytes]
	addlines := strings.Count(jump, "\n")
	sp.line += addlines
	sp.i += utf8.RuneCountInString(jump)
	sp.position += jumpbytes
}

// Find instances of tag IDs.
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

// Find the next HTML tag
func findTag(html string, sp *stringPos, open bool, ids map[string]lineTag) (tag string, selfClosing bool, err error) {
	p := sp.position
	start := strings.IndexAny(html[p:], " >\n")
	if start == -1 {
		return "", false, errors.New("No ending marker found")
	}
	tag = html[p : p+start]
	end := strings.IndexAny(html[p:], ">")
	if end == -1 {
		return "", false, errors.New("No > after <")
	}
	if open {
		selfend := strings.Index(html[p:p+end+1], "/>")
		if selfend > -1 {
			if debug {
				dbgmsg("Self-closing tag\n")
			}
			selfClosing = true
			if html[p+start-1] == '/' {
				tag = html[p : p+start-1]
			}
		}
	}
	// ids as nil is used for end tags.
	if open && start+1 < end {
		remaining := html[sp.position+start+1 : sp.position+end]
		if len(remaining) > 0 {
			findIds(ids, tag, remaining, sp)
		}
	}
	jumpover(html, end, sp)
	tag = strings.ToLower(tag)
	return tag, selfClosing, err
}

// Skip over HTML comments or doctype declarations.
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

// Skip over <script>..</script>.
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

type tagStack []lineTag

func (opentags *tagStack) Pop() (toptag lineTag) {
	n := len(*opentags)
	toptag = (*opentags)[n-1]
	*opentags = (*opentags)[:n-1]
	return toptag
}

func (opentags *tagStack) Push(toptag lineTag) {
	*opentags = append(*opentags, toptag)
}

// Given a closing tag "tag", close it, and any unopened tags, and
// report errors on the unclosed tags.
func (opentags *tagStack) CloseOpenTags(sp stringPos, tag string) (toptag lineTag) {
	tag = strings.ToLower(tag)
	if len(*opentags) == 0 {
		fmt.Printf("%s: there were too many closing tags.\n", sp.String())
		return toptag
	}
	// Pop one tag
	toptag = opentags.Pop()
	if toptag.tag == tag {
		// The tag at the top of the stack was the one we expected, so
		// we have finished without finding an error.
		return toptag
	}
	closed := false
	n := len(*opentags)
	for scrape := n - 1; scrape >= 0; scrape-- {
		scrapeTag := (*opentags)[scrape]
		if scrapeTag.tag == tag {
			fmt.Printf("%s tag mismatch: <%s> </%s>: ",
				sp.String(), toptag.tag, tag)
			unclosed := n - scrape
			fmt.Printf("popping %d unclosed tags.\n", unclosed)
			debugDepth -= unclosed
			for i := 0; i < unclosed; i++ {
				unctag := (*opentags).Pop()
				fmt.Printf("\t%s:%d: <%s> unclosed.\n",
					sp.filename, unctag.line, unctag.tag)
			}
			fmt.Printf("%s:%d: <%s> unclosed.\n",
				sp.filename, toptag.line, toptag.tag)
			closed = true
			break
		}
	}
	if !closed {
		fmt.Printf("%s closing tag </%s> with no opening tag.\n",
			sp.String(), tag)
		opentags.Push(toptag)
	}
	return toptag
}

var debug = false
var debugDepth = 0

func dbgmsg(format string, a ...any) {
	for i := 0; i < debugDepth; i++ {
		fmt.Print("\t")
	}
	fmt.Printf(format, a...)
	fmt.Print("\n")
}

// Validate the HTML in "html", which comes from a file with the name
// "filename".
func validate(html string, filename string) {
	// The stack of currently open tags.
	var opentags tagStack
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
			// Start of an HTML tag
			sp.Add(c)
			c := split[sp.i]
			switch c {
			case "/":

				// The current tag was a closing tag, so pop the stack
				// "opentags" to find a match.
				sp.Add(c)
				tag, _, err := findTag(html, &sp, false, ids)
				if err != nil {
					fmt.Printf("%s error %s\n",
						sp.String(), err)
					break
				}
				if debug {
					debugDepth--
					dbgmsg("Closing </%s>", tag)
				}
				delete(nestTags, tag)
				toptag := opentags.CloseOpenTags(sp, tag)
				if !nonEmpty[tag] {
					break
				}
				var j int
				empty := true
				for j = toptag.position + 1; j < sp.position-2-len(tag); j++ {
					if !unicode.IsSpace(rune(html[j])) {
						empty = false
					}
				}
				if empty {
					fmt.Printf("%s:%d: empty %s tag.\n",
						sp.filename, toptag.line, tag)
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
				// Open tag, consume the whole tag then decide what to
				// do with it.
				tag, selfClosing, err := findTag(html, &sp, true, ids)
				if !valid[tag] {
					fmt.Printf("%s invalid tag <%s>.\n",
						sp.String(), tag)
				}
				if err != nil {
					fmt.Printf("%s error %s\n",
						sp.String(), err)
					break
				}
				if noClose[tag] {
					break
				}
				if selfClosing {
					if debug {
						dbgmsg("Self-closing <%s/>", tag)
					}
					break
				}
				if debug {
					dbgmsg("Opening <%s>", tag)
					debugDepth++
				}
				var lt lineTag
				lt.tag = tag
				lt.line = sp.line
				lt.position = sp.position
				opentags.Push(lt)
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
				if tag == "script" {
					skipScript(html, &sp)
				}
			}
		case "\n":
			// End of a line, increment the line number
			sp.line++
			fallthrough
		default:
			// Add this string to the position.
			sp.Add(c)
		}
	}
	if len(opentags) > 0 {
		fmt.Printf("There are %d unclosed tags:\n", len(opentags))
		for i := 0; i < len(opentags); i++ {
			fmt.Printf("%s:%d: <%s>\n", sp.filename, opentags[i].line,
				opentags[i].tag)
		}
	}
}

func main() {
	debugFlag := flag.Bool("debug", false, "Switch on debugging output")
	flag.Parse()
	debug = *debugFlag
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
