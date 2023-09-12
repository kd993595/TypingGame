package main

import (
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"strings"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	wordsystem "github.com/kd993595/TypingGame/WordSystem"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	whiteImage = ebiten.NewImage(3, 3)
	//colorImage = ebiten.NewImage(3, 3)

	// whiteSubImage is an internal sub image of whiteImage.
	// Use whiteSubImage at DrawTriangles instead of whiteImage in order to avoid bleeding edges.
	whiteSubImage = whiteImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)

	vertexColor = color.RGBA{20, 20, 200, 255}
	normalFont  font.Face
	mapImg      *ebiten.Image
)

const (
	screenWidth   = 1600
	screenHeight  = 1200
	dpi           = 144
	dstFromVertex = 5
)

func init() {
	whiteImage.Fill(color.White)

	dat, err := os.ReadFile("./fonts/AbstractGroovy.ttf") // SuperDream.ttf also works really well
	if err != nil {
		log.Fatal(err)
	}
	tt, err := opentype.Parse(dat)
	if err != nil {
		log.Fatal(err)
	}

	normalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    10,
		DPI:     dpi,
		Hinting: font.HintingVertical,
	})
	if err != nil {
		log.Fatal(err)
	}

	mapImg, _, err = ebitenutil.NewImageFromFile("mapLevel1.png")
	if err != nil {
		log.Fatal(err)
	}
}

type Movable interface {
	MoveTo(Vertex)
}

type Game struct {
	counter int
	runes   []rune
	text    string
	mymap   MapSystem
	wSystem wordsystem.WordSystem
	aSystem ArmySystem
}

type Vec2 struct {
	x float32
	y float32
}

type Vertex struct {
	name             string
	center           Vec2
	adjacentVertices []Vertex
}

func createVertex(name string, x, y float32) Vertex {
	return Vertex{name: name, center: Vec2{x, y}}
}

type MapSystem struct {
	vertices []Vertex
}

func (m *MapSystem) getVertex(name string) *Vertex {
	for i := range m.vertices {
		if m.vertices[i].name == name {
			return &m.vertices[i]
		}
	}
	return nil
}

type Warrior struct {
	station Vertex
	speed   float32
	pos     Vec2
	canMove bool
}

func (w *Warrior) MoveTo(v Vertex) {
	w.station = v
	w.canMove = false
}

func (w *Warrior) checkPos() {
	if !checkDistance(w.pos, w.station.center, 1) {
		dir := unitVector(w.station.center, w.pos)
		displacement := Vec2{x: dir.x * w.speed / float32(ebiten.TPS()*2), y: dir.y * w.speed / float32(ebiten.TPS()*2)}
		w.pos.x = w.pos.x + displacement.x
		w.pos.y = w.pos.y + displacement.y
	} else {
		w.canMove = true
	}
}

func createWarrior(initial Vertex, speed float32, p Vec2) Warrior {
	return Warrior{station: initial, speed: speed, pos: p, canMove: true}
}

type ArmySystem struct {
	warriorTroops []Warrior
}

func (a *ArmySystem) movePlayerWarriors(to, from Vertex) {
	for i := 0; i < len(a.warriorTroops); i++ {
		if a.warriorTroops[i].station.name == from.name && a.warriorTroops[i].canMove {
			a.warriorTroops[i].MoveTo(to)
		}
	}
}

func createMaps() MapSystem {
	vertex1 := createVertex("vertex1", 100, 180)
	vertex2 := createVertex("vertex2", 420, 180)
	vertex3 := createVertex("vertex3", 100, 940)
	vertex4 := createVertex("vertex4", 700, 180)
	vertex5 := createVertex("vertex5", 420, 580)
	vertex6 := createVertex("vertex6", 340, 900)
	vertex7 := createVertex("vertex7", 980, 580)
	vertex8 := createVertex("vertex8", 740, 900)
	vertex9 := createVertex("vertex9", 940, 260)
	vertex10 := createVertex("vertex10", 740, 1100)
	vertex11 := createVertex("vertex11", 980, 860)
	vertex12 := createVertex("vertex12", 1460, 260)
	vertex13 := createVertex("vertex13", 1460, 860)
	vertex14 := createVertex("vertex14", 1500, 1100)
	vertex1.adjacentVertices = []Vertex{vertex2, vertex3}
	vertex2.adjacentVertices = []Vertex{vertex1, vertex4, vertex5}
	vertex3.adjacentVertices = []Vertex{vertex1, vertex6}
	vertex4.adjacentVertices = []Vertex{vertex2, vertex9}
	vertex5.adjacentVertices = []Vertex{vertex2, vertex7, vertex8}
	vertex6.adjacentVertices = []Vertex{vertex3, vertex8}
	vertex7.adjacentVertices = []Vertex{vertex5, vertex9, vertex11, vertex12}
	vertex8.adjacentVertices = []Vertex{vertex5, vertex6, vertex10, vertex11}
	vertex9.adjacentVertices = []Vertex{vertex4, vertex7, vertex12}
	vertex10.adjacentVertices = []Vertex{vertex8, vertex14}
	vertex11.adjacentVertices = []Vertex{vertex7, vertex8, vertex13}
	vertex12.adjacentVertices = []Vertex{vertex7, vertex9, vertex13}
	vertex13.adjacentVertices = []Vertex{vertex11, vertex12, vertex14}
	vertex14.adjacentVertices = []Vertex{vertex10, vertex13}

	newmap := MapSystem{vertices: []Vertex{vertex1, vertex2, vertex3, vertex4, vertex5, vertex6, vertex7, vertex8, vertex9, vertex10, vertex11, vertex12, vertex13, vertex14}}
	return newmap
}

//Draws a line between two points given both as Vector2
func drawLine(dst *ebiten.Image, to, from Vec2, aa bool) {
	var path vector.Path

	path.MoveTo(from.x, from.y)
	path.LineTo(to.x, to.y)
	path.Close()

	var vs []ebiten.Vertex
	var is []uint16

	op := &vector.StrokeOptions{}
	op.Width = 2
	op.LineJoin = vector.LineJoinRound
	vs, is = path.AppendVerticesAndIndicesForStroke(nil, nil, op)

	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = 0xdb / float32(0xff)
		vs[i].ColorG = 0x56 / float32(0xff)
		vs[i].ColorB = 0x20 / float32(0xff)
		vs[i].ColorA = 1
	}

	op2 := &ebiten.DrawTrianglesOptions{}
	op2.AntiAlias = aa
	dst.DrawTriangles(vs, is, whiteSubImage, op2)
}

//draws a square at given Vector2 point with given size and color
func drawSquare(dst *ebiten.Image, center Vec2, size float32, clr color.Color, aa bool) {
	vector.DrawFilledRect(dst, center.x-(size/2), center.y-(size/2), size, size, clr, aa)
}

func drawCircle(dst *ebiten.Image, center Vec2, size float32, clr color.Color, aa bool) {
	vector.DrawFilledCircle(dst, center.x-(size/2), center.y-(size/2), size, clr, aa)
}

func removeWhitespaceCharacters(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func repeatingKeyPressed(key ebiten.Key) bool {
	const (
		delay    = 30
		interval = 3
	)
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	if d >= delay && (d-delay)%interval == 0 {
		return true
	}
	return false
}

func (g *Game) Update() error {
	g.runes = ebiten.AppendInputChars(g.runes[:0])
	g.text += strings.ToLower(removeWhitespaceCharacters(string(g.runes)))
	clearText, entityMatched := g.wSystem.CheckChars(g.text)
	if clearText {
		g.text = ""
		if strings.HasPrefix(entityMatched, "vertex") {
			g.wSystem.RemoveWordEntities("subvertex")
			t_vertex := g.mymap.getVertex(entityMatched)
			for i := range t_vertex.adjacentVertices {
				dirvec := makeHalfVector(t_vertex.center, t_vertex.adjacentVertices[i].center)
				g.wSystem.CreateWordEntity("subvertex"+t_vertex.adjacentVertices[i].name+"&"+t_vertex.name, int(t_vertex.center.x+dirvec.x), int(t_vertex.center.y+dirvec.y)+10)
			}
		} else if strings.HasPrefix(entityMatched, "subvertex") {
			tmp := entityMatched[len("subvertex"):]
			t_vertices := strings.Split(tmp, "&")
			g.aSystem.movePlayerWarriors(*g.mymap.getVertex(t_vertices[0]), *g.mymap.getVertex(t_vertices[1]))
			g.wSystem.RemoveWordEntities("subvertex")
		}

	}

	if repeatingKeyPressed(ebiten.KeyBackspace) {
		if len(g.text) >= 1 {
			g.text = g.text[:len(g.text)-1]
		}
	}

	for i := range g.aSystem.warriorTroops {
		g.aSystem.warriorTroops[i].checkPos()
	}

	g.counter++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	//screen.Fill(color.RGBA{0, 163, 57, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(2.5, 2.5)
	screen.DrawImage(mapImg, op)

	//drawLine(screen, Vec2{100, 100}, Vec2{300, 300}, true)
	//drawSquare(screen, Vec2{100, 100}, 20, color.RGBA{0, 0, 255, 255}, true)
	for _, v := range g.mymap.vertices {
		drawSquare(screen, v.center, 20, vertexColor, true)
		for _, z := range v.adjacentVertices {
			drawLine(screen, z.center, v.center, true)
		}
	}

	for _, v := range g.wSystem.WordComponents {
		text.Draw(screen, v.WordChar, normalFont, v.X-20, v.Y-20, color.White) //gotta make position of word dynamic to vertex
		if v.Highlighted {
			text.Draw(screen, g.text, normalFont, v.X-20, v.Y-20, color.RGBA{255, 255, 0, 255})
		}
	}

	for _, v := range g.aSystem.warriorTroops {
		drawCircle(screen, v.pos, 5, color.RGBA{185, 153, 2, 255}, true)
	}

	text.Draw(screen, g.text, normalFont, 0, 15, color.RGBA{200, 20, 20, 255})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	var g Game
	g.mymap = createMaps()
	g.wSystem = wordsystem.InitializeWordSystem("words.txt")
	g.aSystem = ArmySystem{warriorTroops: []Warrior{}}
	g.aSystem.warriorTroops = append(g.aSystem.warriorTroops, createWarrior(*g.mymap.getVertex("vertex2"), 5, g.mymap.getVertex("vertex2").center))
	g.aSystem.warriorTroops = append(g.aSystem.warriorTroops, createWarrior(*g.mymap.getVertex("vertex1"), 5, g.mymap.getVertex("vertex1").center))
	g.aSystem.warriorTroops = append(g.aSystem.warriorTroops, createWarrior(*g.mymap.getVertex("vertex3"), 5, g.mymap.getVertex("vertex3").center))

	for i := range g.mymap.vertices {
		thisVertex := g.mymap.vertices[i]
		g.wSystem.CreateWordEntity(thisVertex.name, int(thisVertex.center.x), int(thisVertex.center.y))
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Typing Game")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.MaximizeWindow()
	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}

/*func lerp(a, b, c float32) float32 {
	return a*(1-c) + b*c
}*/

func unitVector(to, from Vec2) Vec2 {
	const multiplier = 50
	v1 := Vec2{to.x - from.x, to.y - from.y}
	mag := math.Sqrt(float64(v1.x*v1.x + v1.y*v1.y))
	v1.x = (v1.x / float32(mag)) * multiplier
	v1.y = (v1.y / float32(mag)) * multiplier
	return v1
}

func makeHalfVector(to, from Vec2) Vec2 {
	//const multiplier = 50
	v1 := Vec2{(from.x - to.x) / 3, (from.y - to.y) / 3}
	return v1
}

//returns true if distance between 2 vectors given is less than distance argument
func checkDistance(a, b Vec2, dist int) bool {
	t := math.Pow(float64(a.x-b.x), 2) + math.Pow(float64(a.y-b.y), 2)
	dist = dist * dist
	return t < float64(dist)
}
