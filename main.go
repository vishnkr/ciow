package main

import (
	"errors"
	"image/color"
	"log"
	"math/rand"
	"os"

	st "github.com/golang-collections/collections/stack"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/exp/slices"
)


type chip8 struct{
	memory [4096]byte
	pc uint16
	registers [16]byte
	i uint16 
	sound_timer byte
	delay_timer byte
	stack *st.Stack
	screen [32][64]bool
	opcodes map[uint16]opHandler
}

type opHandler func(ins uint16)

var fontset []byte= []byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

var KEY_MAP = map[byte]ebiten.Key{1:ebiten.Key1, 2: ebiten.Key2, 3: ebiten.Key3, 0xc: ebiten.Key4, 4: ebiten.KeyQ, 5: ebiten.KeyW, 6: ebiten.KeyE, 0xd: ebiten.KeyR, 7: ebiten.KeyA, 8: ebiten.KeyS, 9: ebiten.KeyD, 0xe: ebiten.KeyF, 0xa: ebiten.KeyZ, 0: ebiten.KeyX, 0xb: ebiten.KeyC, 0xf:ebiten.KeyV}
var REVERSE_KEY_MAP = map[ebiten.Key]byte{ebiten.Key1:1,ebiten.Key2:2,ebiten.Key3:3, ebiten.Key4:0xc, ebiten.KeyQ: 4, ebiten.KeyW:5, ebiten.KeyE:6, ebiten.KeyR:0xd, ebiten.KeyA:7, ebiten.KeyS:8, ebiten.KeyD:9, ebiten.KeyF:0xe, ebiten.KeyZ:0xa,  ebiten.KeyX:0,  ebiten.KeyC:0xb, ebiten.KeyV:0xf}
func extractValues(instruction uint16) (opcode, x, y, kk, n byte, nnn uint16) {

	opcode = byte((instruction & 0xf000) >> 12)
	x = byte((instruction & 0x0F00) >> 8)
	y = byte((instruction & 0x00F0) >> 4)
	kk = byte(instruction & 0x00FF)
	n = byte(instruction & 0x000F)
	nnn = instruction & 0x0FFF

	return opcode, x, y, kk, n, nnn
}

func (c *chip8) exec(ins uint16){
	op, x, y, kk, n, nnn := extractValues(ins)
	//var handler opHandler = c.opcodes[uint16(opCode)]
	//handler(ins)

	switch op{
	case 0:
		switch kk{
		case 0xe0:
			for i := range c.screen {
				for j := range c.screen[i] {
					c.screen[i][j] = false
				}
			}
			break
		case 0xee:
			val := c.stack.Pop()
			if val!=nil{
				c.pc = val.(uint16)
			}
			break
		}
		break
	case 1:
		c.pc = nnn
		break
	case 2:
		c.stack.Push(c.pc)
		c.pc = nnn
		break
	case 3:
		if c.registers[x]==kk{
			c.pc+=2
		}
		break
	case 4:
		if c.registers[x]!=kk{
			c.pc+=2
		}
		break
	case 5:
		if c.registers[x]==c.registers[y]{
			c.pc+=2
		}
		break
	case 6:
		c.registers[x]=kk
		break
	case 7:
		c.registers[x] += kk
		break
	case 8:
		var vf byte
		switch n{
		case 0:
			c.registers[x] = c.registers[y]
			break
		case 1:
			c.registers[x] = c.registers[x] | c.registers[y]
			break
		case 2:
			c.registers[x] = c.registers[x] & c.registers[y]
			break
		case 3:
			c.registers[x] = c.registers[x] ^ c.registers[y]
			break
		case 4:
			if uint16(c.registers[x])+uint16(c.registers[y]) > 255 {
				vf = 1
			} else{
				vf = 0
			}
			c.registers[x] = c.registers[x]+c.registers[y]
			c.registers[0xf]=vf
			break
		case 5:
			if c.registers[x]>c.registers[y] {
				vf = 1
			} else{
				vf = 0
			}
			c.registers[x] = c.registers[x] - c.registers[y]
			c.registers[0xf]=vf
			break
		case 6:
			c.registers[0xf] = c.registers[x] & 1
			c.registers[x] >>= 1
			break
		case 7:
			if c.registers[x] > c.registers[y]{
				c.registers[0xf] = 0
			} else{
				c.registers[0xf] = 1
			}
			c.registers[x] = c.registers[y] - c.registers[x]
			break
		case 0xe:
			c.registers[0xf] = (c.registers[x] >> 7) & 1
			c.registers[x] <<= 1
			
			break
		}
		break
	case 9:
		if c.registers[x]!=c.registers[y]{
			c.pc+=2
		}
		break
	case 0xa:
		c.i = nnn
		break
	case 0xb:
		c.pc = nnn+uint16(c.registers[0])
		break
	case 0xc:
		c.registers[x] = byte(rand.Float32()*255) & kk
		break
	case 0xd:
		c.draw(x,y,n)
		break
	case 0xe:
		if c.registers[x]<0 || c.registers[x]> 0xf{ break}
		if (kk==0x9e && ebiten.IsKeyPressed(KEY_MAP[c.registers[x]]) || kk==0xa1 && !ebiten.IsKeyPressed(KEY_MAP[c.registers[x]])){
			c.pc+=2
		}
		break
	case 0xf:
		switch kk{
		case 0x07:
			c.registers[x]=c.delay_timer
		case 0x0a:
			keys := make([]ebiten.Key,16)
			for len(keys)==0{
				keys = inpututil.AppendJustPressedKeys(keys)
			}
			c.registers[x] = byte(REVERSE_KEY_MAP[keys[0]])
			break
		case 0x15:
			c.delay_timer = c.registers[x]
			break
		case 0x18:
			c.sound_timer = c.registers[x]
			break
		case 0x1e:
			c.i = c.i + uint16(c.registers[x])
			break
		case 0x29:
			c.i = uint16(0x50+5*(c.registers[x]&0xf)) & 0xff
			break
		case 0x33:
			num := c.registers[x]
			c.memory[c.i] = num / 100
			c.memory[c.i+1]= (num%100)/10
			c.memory[c.i+2] = num%10
			break
		case 0x55:
			var b byte
			for b = 0; b <= x; b++ {
				c.memory[c.i+uint16(b)] = c.registers[b]
			}
			break
		case 0x65:
			var b byte
			for b = 0; b <= x; b++ {
				c.registers[b] = c.memory[c.i+uint16(b)] 
			}
			break
		default:
			break
		}
	default:
		break
	}

	if c.delay_timer>0{
		c.delay_timer-=1
	}
	if c.sound_timer>0{
		c.sound_timer-=1
	}
}

func (c *chip8) draw(x,y,n byte){
	vx,vy := uint16(c.registers[x]%64), uint16(c.registers[y]%32)
	var i, j uint16
	var vfFlag = false
	for i=0; i<uint16(n);i++{
		for j=0;j<8;j++{
			b:= c.memory[c.i+i] >> (8-j-1) & 0x1
			if b==0 || vx+j>64 || vy+i>32{
				continue
			}
			if c.screen[vy+i][vx+j]{
				vfFlag= true
			}
			c.screen[vy+i][vx+j] = !c.screen[vy+i][vx+j]
		}
	}
	if vfFlag{
		c.registers[0xf] = 1
	} else { c.registers[0xf] = 0}
}

type Game struct{
	chip8
}

func (g *Game) Update() error {
	keys:= make([]ebiten.Key,16)
	keys = inpututil.AppendJustPressedKeys(keys)
	if slices.Contains(keys,ebiten.KeyC) && slices.Contains(keys,ebiten.KeyControl){
		return errors.New("terminated")
	}
	
		instruction := g.chip8.memory[g.chip8.pc: g.chip8.pc+2]
		g.chip8.pc+=2
		ins:= uint16(instruction[0])<<8 | uint16(instruction[1])
		log.Printf("%x [INS] %x [PC]", ins, g.chip8.pc)
		
		g.chip8.exec(uint16(ins))
	
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for i :=0; i<32; i++{
		for j:=0;j<64;j++{
			if g.chip8.screen[i][j]{
				screen.Set(j,i,color.White)
			}
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 64,32
}

func newChip8(filePath string)chip8{
	data, err := os.ReadFile(filePath)
	if err!=nil{
		panic(err)
	}
	var new chip8 = chip8{
		pc: 0x200,
		i:0,
		stack: st.New(),
		sound_timer: 0,
		delay_timer: 0,    
	}
	copy(new.memory[0x200:],data)
	copy(new.memory[0x50:0x9f],fontset)
	return new
}
func main(){
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Chip8 emulator")
	var path string
	if len(os.Args) > 1 {
		path = os.Args[1]
	} else {
		panic("File path not provided")
	}
	game:= Game{
		chip8: newChip8(path),
	}
	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
	
}