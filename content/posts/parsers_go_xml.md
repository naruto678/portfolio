---
title: "Writing xml parser from scratch in golang"
date: 2023-01-09T02:31:54+05:30
---

### Introduction 
When we think about lexers, parsers , interpreters , we tend to conjure up fairly complex string processing and some 
black magic under the hood. To be honest, tools like **LLVM and ANTLR** are super-duper complex under the hood <take this with a grain of salt> , and the cover of [Compilers , Principles and Tools(The Dragon book)](https://www-2.dc.uba.ar/staff/becher/dragon.pdf) add fuel to the fire.  It shows a knight trying to slay a dragon on the cover . 

this is my meagre attempt to slay a baby caterpillar (**parsing xml**) 

Today we will write a [Recursive Descent Parser](https://en.wikipedia.org/wiki/Recursive_descent_parser) . Why this, you may ask , when there are other types of parsers ? Well , I  felt this was very intutive to implement  once you got the underlying CFG(Context Free Grammar) correct, and this is my way of learning about the underlying machinery of some better compiler anyway . 

As Feynman said "If you cannot build it , you do not understand it " 🙂

So enough philosophical talk , let's get down to business .

In this series we will build a fairly simple xml parser and not a fully-fledged one like ([JAXB](https://docs.oracle.com/javase/tutorial/jaxb/intro/index.html) or [SAX](https://www.baeldung.com/java-sax-parser)). We will slowly ramp up the grammar and hopefully we can get a bit close to the original spec 


 First I would like to give a bit of a tutorial of BNF(Backus Naur Form) and CFG and some of the terms like lexemes, tokens and then we can get started with the implementation . if you are aware of these terms , feel free to jump around. 

### Some prerequisite ideas and concepts and a quick refresher
------------------------------------------
- The first step that we must take while writing a parser/compiler/interpreter is scanning/lexing or "lexical analysis". A lexer takes in a stream of characters from the input file and converts them into groups of words called tokens .
- These tokens are passed as input to a parser that takes these tokens which finally builds the ParseTree/ AST(Abstract Syntax Tree) . This is the tree that the interpreter walks through to interpret expressions/statements 


### Tokens vs Lexemes 
--------------------------------------------------------------------
A lexer scans through each character in the input source file and then chunks them into groups of words called lexemes . 

Whereas you can imagine a token being something like this : 

{{< highlight go "linenos=table,linenostart=1" >}}
type Token struct{
	Lexeme string 
	TokenType TokenType
}

type TokenType byte 

{{< / highlight >}}

### Context Free Grammar and Backus-Naur-Form
-----------------------------------------------
A grammar in the layman sense of the word is basically a collection of symbols(called alphabet)  and it defines rules for generating sentences that are acceptable and correct . In our case these symbols are the tokens that are emitted by the lexer . 

The way to define those rules leads us to to the Backus-Naur-Form or in this case (Extended BNF) 
```
digit -> 0 | 1| 2| 3| 4| 5| 6 |7 | 8 | 9  
natural-num -> natural-num | digit
```

Here every line is what we refer as a production. Every production must start with a **head** which is its name and **body** which defines the rule . 

Every rule can either refer itself or it can be a terminal ie a number or a string literal  . 

As you can guess , this is a very concise way of representing all natural numbers and we can generate any natural number from these two rules alone. 

### Interpretation of the Extended BNF to code 
----------------------
- every rule becomes a method name 
- for every {}.* there should be a while loop 
- and every | creates a if-else check inside the method 


### Let's Dive in . 
----------------------------------------
#### Writing the Lexer 

First we identify things that we want to be called as lexemes . for a simple xml parsers like ours 
we have 
```
	< LPAREN
	> RPAREN
	/ SLASH
	* STRING
	
```
In addition to this we would have a **EOF** token type to mark the end of our parsing 

{{< highlight go "linenos=table,linenostart=1" >}}

type TokenType byte

const (
	LPAREN TokenType = '<'
	RPAREN TokenType = '>'
	SLASH  TokenType = '/'
	STRING TokenType = '*'
)

//Takes a tokentype and a value
func NewToken(tokenType TokenType, value string, name string) Token {
	return Token{
		TokenType: tokenType,
		Literal:   value,
		Name:      name,
	}
}

type Token struct {
	TokenType
	Literal string
	Name    string
}

// will be useful later while printing
func (t *Token) String() string {
	return fmt.Sprintf("Token(%s, %s)", t.Name, t.Literal)
}

// the special EOF token 
func EOF() Token {
	return Token{
		Literal: "EOF",
		Name:    "EOF",
	}
}

{{< / highlight >}}

Here we have defined utility functions for creating tokens . Let's go to the lexer
{{< highlight go "linenos=table,linenostart=1" >}}

func Lexer(source string) []Token {
	var tokens []Token
	current_pos := 0
	for current_pos < len(source) { 
		// read character by character to the end of file 
		switch source[current_pos] {
		case byte(LPAREN):
			tokens = append(tokens, NewToken(LPAREN, "<", "LPAREN"))
			current_pos += 1
		case byte(RPAREN):
			tokens = append(tokens, NewToken(RPAREN, ">", "RPAREN"))
			current_pos += 1
		case byte(SLASH):
			tokens = append(tokens, NewToken(SLASH, "/", "SLASH"))
			current_pos += 1
		case byte('\n'), byte(' '), byte('\t'): 
		// ignore whitespace, tabs and new lines
			current_pos++
		default:
			literal := ""
			for current_pos < len(source) && isLetter(source[current_pos]) {
				literal += string(source[current_pos])
				current_pos++
			}
			tokens = append(tokens, NewToken(STRING, literal, "STRING"))
		}
	}
	tokens = append(tokens, EOF())
	return tokens
}

func isLetter(letter byte) bool {
	result := letter == '<' || letter == '>'
	return !result
}

{{< / highlight >}}

In the above function the lexer goes through character by character and emits tokens that would be consumed by the parser 

Writing some tests is always a good idea . 

{{< highlight go "linenos=table,linenostart=1" >}}
func TestLexer(t *testing.T) {
	xml_source := "<A>FirstTest</A>"
	tokens := Lexer(xml_source)
	for _, token := range tokens {
		fmt.Println(token.String())
	}
	xml_source = `
		<A>
			<B>some_value</B>
			<C>some_second_value</C>
			<D>some_third_value</D>
			<E> some_more_value </E>
			<F>
				<G>some_nested_value</G>
				<H>some_more_nesting</H>
			</F>
		</A>
	`
	tokens = Lexer(xml_source)
	for _, token := range tokens {
		fmt.Println(token.String())
	}

}

{{< / highlight >}}

On running the tst we see the tokens that are getting emitted by the lexer. We are now ready to start with the parser and create AST's. But before we get started on that we need to define some sort of grammar for our xml parser . 

In my attempt to keep the grammar as simple as possible , this is what I came up with 
```
node -> <string>string</string> | <string>{node}+</string> // since we can have multiple nested nodes inside a node 
```
Since we do not care about having stuff like attributes inside our tag , this should suffice and we can add new production rules later down the line when we want to enrich the same 

Keeping the above rules in mind , we can start with a skeleton of something like this 

{{< highlight go "linenos=table,linenostart=1" >}}

type Parser struct{
	// some fields 
}

func (p *Parser) Parse() {
}

func (p *Parser) Node() {
	/* since we have '|' and {}+ in the production rule we 
	would have to have a if-else check and 
	we would need to have a while loop and then 
	inside each we  would call the p.Node method recurisively . 
	*/
}


{{< / highlight >}}
Let's start fleshing out the details 

We define a xml node as follows 

{{< highlight go "linenos=table,linenostart=1" >}}
type Node struct {
	Name      string
	Parent    *Node
	Children  []*Node
	TextValue string
}

func NewNode(Name string, Parent Node) Node {
	return Node{
		Name:     Name,
		Parent:   &Parent,
		Children: []*Node{},
	}
}

{{< / highlight >}}

Create a new parser 

{{< highlight go "linenos=table,linenostart=1" >}}
type Parser struct {
	Idx    int
	Tokens []Token
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		Idx:    0,
		Tokens: tokens,
	}

}

func (p *Parser) Unmarshall() *Node {
	return nil
}

func (p *Parser) Marshall() (string, error) {
	return "", fmt.Errorf("not implemented yet")
}

func (p *Parser) Increment() {
	p.Idx++
}

func (p *Parser) Peek() Token {
	return p.Tokens[p.Idx]
}

func (p *Parser) GetCurrentToken() Token {
	token_val := p.GetToken(p.Idx)
	p.Increment()
	return token_val
}

func (p *Parser) PeekNext() Token {
	if p.Idx+1 >= len(p.Tokens) {
		return EOF()
	} else {
		return p.Tokens[p.Idx+1]
	}
}

func (p *Parser) GetToken(idx int) Token {
	if p.Idx > len(p.Tokens) {
		return EOF()
	}
	return p.Tokens[p.Idx]
}

func (p *Parser) Parse() (Node, error) {
	parent_node := Node{}
	return p.Node(parent_node)
}

{{< / highlight >}}

The parse method is the place where we put the logic for creating the parse tree/AST

{{< highlight go "linenos=table,linenostart=1" >}}
func (p *Parser) Node(parent Node) (Node, error) {
var curr_node Node
if p.GetCurrentToken().TokenType == LPAREN {
	if p.Peek().TokenType == STRING {
		curr_node = NewNode(p.GetCurrentToken().Literal, parent)
		if curr_token := p.GetCurrentToken(); curr_token.TokenType != RPAREN {
			return curr_node, fmt.Errorf("expected '>' Got %s", p.Peek().Name)
		}
		next_token := p.Peek()

		for next_token.TokenType == LPAREN && \n
		p.PeekNext().TokenType != SLASH {
			child_node, err := p.Node(curr_node)
			if err != nil {
				return curr_node, err
			}
			curr_node.Children = append(curr_node.Children, &child_node)
			next_token = p.Peek()

		}
		if next_token.TokenType == STRING {
			curr_node.TextValue = next_token.Literal
			p.GetCurrentToken()
		}
		if p.GetCurrentToken().TokenType == LPAREN &&
		 p.GetCurrentToken().TokenType == SLASH {
			closing_symbol_name := p.GetCurrentToken()
			if curr_node.Name != closing_symbol_name.Literal {
				return curr_node, fmt.Errorf("closing tag does not match .
		expected [%s] got [%s]", curr_node.Name, closing_symbol_name.Literal)
			}
			if p.GetCurrentToken().TokenType != RPAREN {
				return curr_node, fmt.Errorf("expected > got [%s]",
				 p.Peek().Literal)
			}
		}

		return curr_node, nil
	}
}
return curr_node, fmt.Errorf("expected < . got %s", p.Peek().Name)
}

{{< / highlight >}}

Writing a printer for the AST
As always writing tests is a good idea 
------------
test file 
```
<A>
	<B> The value of B</B>
	<C> The value of C</C>
	<D> the value of D</D>
	<E>some value of E</E>
	<F>
		<G> the value of G</G>
		<B> The value of B</B>
		<C> The value of C</C>
		<D> the value of D</D>
		<E>some value of E</E>
	</F>
</A> 

```

{{< highlight go "linenos=table,linenostart=1" >}}

func (p *Parser) Print(printer Printer) {
	printer.Print()
}
type Printer interface {


	Print()
}

type ASTPrinter struct {
	Root Node
}

func (p ASTPrinter) Print() {

}

type BFSPrinter struct {
	Root Node
}
// Breadth first Printer
func (p BFSPrinter) Print() {
	queue := []Node{}
	queue = append(queue, p.Root)
	println(p.Root.Name)
	for len(queue) > 0 {
		result := ""
		for size := len(queue); size > 0; size-- {
			top := queue[0]
			queue = queue[1:]
			for _, child := range top.Children {
				result += fmt.Sprintf("%s-%s", child.Name, child.TextValue) + " "
				queue = append(queue, *child)
			}

		}
		fmt.Println(result)
	}
}

func TestParser(t *testing.T) {
	content, err := os.ReadFile("./resources/test/simple.xml")
	if err != nil {
		t.Fatal("Cannot read file ", err.Error())
	}
	source := string(content)

	start_time := time.Now()

	tokens := Lexer(source)

	if len(tokens) == 0 {
		t.Fatal("Lexer failed")
	}

	parser := NewParser(tokens)
	root_node, err := parser.Parse()

	if err != nil {
		t.Fatal(err.Error())
	}

	fmt.Printf("completed parsing in %d ns\n", 
		time.Since(start_time).Nanoseconds())

	bfs_printer := BFSPrinter{
		Root: root_node,
	}
	parser.Print(bfs_printer)
}

{{</highlight>}}
if you reached till here , Give yourself a pat on the back . we got a lot covered in this chapter  In the next chapter we are going to enrich the grammar 
and add some fun methods on the xml node like getNodebyName or some other methods present in the **jaxb unmarshaller**

Thanks for reading. All code avaialble in the repo [xml-parser](https://github.com/naruto678/xml-parser)
