package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// skipEncodeChars are characters that should NOT be Ghost Bits encoded.
// Only / \ = & ? are kept as-is; everything else gets encoded.
var skipEncodeChars = map[rune]bool{
	'/':  true,
	'\\': true,
	'=':  true,
	'&':  true,
	'?':  true,
}

// preferredMap maps low-byte values to preferred Ghost Bits Unicode characters
// sourced from the Cast Attack research (Black Hat ASIA 2026).
// Each rune's low 8 bits MUST equal the key byte.
var preferredMap = map[byte]rune{
	0x0A: '瘊', // \n → U+760A (low byte 0x0A ✓)
	0x0D: '瘍', // \r → U+760D (low byte 0x0D ✓)
	0x25: '严', // % → U+4E25 (low byte 0x25 ✓)
	0x2E: '阮', // . → U+962E (low byte 0x2E ✓)
	0x30: '丰', // 0 → U+4E30 (low byte 0x30 ✓)
	0x32: '甲', // 2 → U+7532 (low byte 0x32 ✓)
	0x33: '⑳', // 3 → U+3233 (low byte 0x33 ✓)
	0x61: 'ᙡ', // a → U+1661 (low byte 0x61 ✓)
	0x63: '㹣', // c → U+3E63 (low byte 0x63 ✓)
	0x65: '来', // e → U+6765 (low byte 0x65 ✓)
	0x6A: '陪', // j → U+966A (low byte 0x6A ✓)
	0x6C: '౬', // l → U+0C6C (low byte 0x6C ✓)
	0x75: '灵', // u → U+7075 (low byte 0x75 ✓)
}

// encodeToGhostBits converts text to Ghost Bits Unicode characters.
// Only / \ = & ? are kept as-is; all other characters are encoded.
func encodeToGhostBits(input string) string {
	var result []rune
	for _, r := range input {
		// Skip encoding for / \ = & ?
		if skipEncodeChars[r] {
			result = append(result, r)
			continue
		}
		// Encode everything else using Ghost Bits
		lowByte := byte(r & 0xFF)
		if ghost, ok := preferredMap[lowByte]; ok {
			result = append(result, ghost)
		} else {
			// Systematic mapping: CJK Unified Ideographs
			result = append(result, rune(0x4E00+int(lowByte)))
		}
	}
	return string(result)
}

// decodeFromGhostBits extracts the low 8 bits from each Unicode character,
// simulating Java's (byte)char truncation behavior.
func decodeFromGhostBits(input string) string {
	var result []byte
	for _, r := range input {
		result = append(result, byte(r&0xFF))
	}
	return string(result)
}

// analyzeGhostBits produces a detailed per-character mapping analysis.
func analyzeGhostBits(input string) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("%-6s  %-12s  %-8s  %-8s  %s", "字符", "Unicode", "低8位", "ASCII", "说明"))
	lines = append(lines, strings.Repeat("─", 58))

	for _, r := range input {
		lowByte := byte(r & 0xFF)
		var asciiRepr string
		switch {
		case lowByte >= 0x20 && lowByte <= 0x7E:
			asciiRepr = fmt.Sprintf("'%c'", lowByte)
		case lowByte == 0x0A:
			asciiRepr = `'\n'`
		case lowByte == 0x0D:
			asciiRepr = `'\r'`
		case lowByte == 0x09:
			asciiRepr = `'\t'`
		default:
			asciiRepr = fmt.Sprintf("0x%02X", lowByte)
		}

		note := ""
		if r > 0xFF {
			note = "Ghost Bit (高位截断)"
		} else if r <= 0x7F {
			note = "纯 ASCII"
		} else {
			note = "扩展 ASCII"
		}

		line := fmt.Sprintf("%-6c  U+%04X       0x%02X      %-8s  %s", r, r, lowByte, asciiRepr, note)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func main() {
	a := app.New()
	w := a.NewWindow("Ghost Bits 转换器")
	w.Resize(fyne.NewSize(680, 580))

	// --- 输入区 ---
	input := widget.NewEntry()
	input.MultiLine = true
	input.Wrapping = fyne.TextWrapWord
	input.SetPlaceHolder("在此输入要转换的文本...\n\n支持任意 Unicode 字符，包括中文、日文、特殊符号等")

	// --- 输出区 ---
	output := widget.NewEntry()
	output.MultiLine = true
	output.Wrapping = fyne.TextWrapWord
	output.SetPlaceHolder("转换结果将显示在此处...")

	// --- 模式选择 ---
	currentMode := "encode"
	modeSelect := widget.NewRadioGroup(
		[]string{"编码 (→Ghost Bits)", "解码 (→低8位)", "分析 (逐字符映射)"},
		func(value string) {
			currentMode = value
		},
	)
	modeSelect.Horizontal = true
	modeSelect.SetSelected("编码 (→Ghost Bits)")

	// --- 转换按钮 ---
	convertBtn := widget.NewButton("🔄 转 换", func() {
		text := input.Text
		if strings.TrimSpace(text) == "" {
			output.SetText("请先在上方输入文本")
			return
		}

		var result string
		switch currentMode {
		case "编码 (→Ghost Bits)":
			result = encodeToGhostBits(text)
		case "解码 (→低8位)":
			result = decodeFromGhostBits(text)
		case "分析 (逐字符映射)":
			result = analyzeGhostBits(text)
		default:
			result = encodeToGhostBits(text)
		}
		output.SetText(result)
	})
	convertBtn.Importance = widget.HighImportance

	// --- 布局 ---
	// 上方输入区
	inputLabel := widget.NewLabelWithStyle("📝 输入文本：", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	inputScroll := container.NewScroll(input)
	inputSection := container.NewBorder(inputLabel, nil, nil, nil, inputScroll)

	// 下方输出区
	outputLabel := widget.NewLabelWithStyle("📤 输出结果：", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	outputScroll := container.NewScroll(output)
	outputSection := container.NewBorder(outputLabel, nil, nil, nil, outputScroll)

	// 中间控制区
	controls := container.NewVBox(
		modeSelect,
		convertBtn,
		widget.NewSeparator(),
	)

	// 整体布局：上下两个区域通过 VSplit 平分，中间控制区固定
	// 使用嵌套 Border：上半部分=inputSection, 中间=controls, 下半部分=outputSection
	split := container.NewVSplit(inputSection, outputSection)
	split.SetOffset(0.5)

	content := container.NewBorder(nil, controls, nil, nil, split)

	// 底部说明
	footer := widget.NewLabel("Ghost Bits：Java char(16位) → byte(8位) 截断，高8位丢弃，仅保留低8位 | 支持任意 Unicode 输入")
	footer.TextStyle = fyne.TextStyle{Italic: true}

	finalContent := container.NewBorder(nil, footer, nil, nil, content)

	w.SetContent(finalContent)
	w.ShowAndRun()
}
