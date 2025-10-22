package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// LaTeXConverter mengelola konversi teks ke LaTeX
type LaTeXConverter struct {
	specialChars map[string]string
	mathPatterns []*regexp.Regexp
}

// NewLaTeXConverter membuat instance baru LaTeXConverter
func NewLaTeXConverter() *LaTeXConverter {
	specialChars := map[string]string{
		"&":  `\&`,
		"%":  `\%`,
		"$":  `\$`,
		"#":  `\#`,
		"_":  `\_`,
		"{":  `\{`,
		"}":  `\}`,
		"~":  `\textasciitilde{}`,
		"^":  `\textasciicircum{}`,
		"\\": `\textbackslash{}`,
	}

	mathPatterns := []*regexp.Regexp{
		regexp.MustCompile(`[=+\-*/^()\[\]]`),
		regexp.MustCompile(`\d+[+\-*/]\d+`),
		regexp.MustCompile(`[a-zA-Z]\s*=\s*\d+`),
		regexp.MustCompile(`[a-zA-Z]\([^)]+\)`),
		regexp.MustCompile(`\\[a-zA-Z]+`),
		regexp.MustCompile(`\$\$`),
		regexp.MustCompile(`\b(sin|cos|tan|log|ln|lim|sum|prod|int)\b`),
	}

	return &LaTeXConverter{
		specialChars: specialChars,
		mathPatterns: mathPatterns,
	}
}

// ConvertToLatex mengkonversi teks biasa ke format LaTeX
func (l *LaTeXConverter) ConvertToLatex(text string) string {
	if text == "" {
		return ""
	}

	lines := strings.Split(text, "\n")
	var latexLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Escape karakter khusus
		escapedLine := l.escapeSpecialChars(line)

		// Deteksi dan format matematika
		if l.isMathExpression(escapedLine) {
			// Process math expressions
			processedMath := l.processMathExpressions(escapedLine)
			latexLines = append(latexLines, fmt.Sprintf("\\[ %s \\]", processedMath))
		} else if l.isInlineMath(escapedLine) {
			// Inline math
			processedMath := l.processMathExpressions(escapedLine)
			latexLines = append(latexLines, fmt.Sprintf("\\( %s \\)", processedMath))
		} else {
			// Regular text
			latexLines = append(latexLines, escapedLine)
		}
	}

	return strings.Join(latexLines, "\n")
}

// escapeSpecialChars mengescape karakter khusus LaTeX
func (l *LaTeXConverter) escapeSpecialChars(text string) string {
	result := text
	for char, replacement := range l.specialChars {
		result = strings.ReplaceAll(result, char, replacement)
	}
	return result
}

// isMathExpression mengecek apakah teks adalah ekspresi matematika
func (l *LaTeXConverter) isMathExpression(text string) bool {
	if strings.Contains(text, "\\") {
		return true
	}

	// Check if it's mostly mathematical symbols
	mathCharCount := 0
	totalChars := 0

	for _, char := range text {
		if unicode.IsSpace(char) {
			continue
		}
		totalChars++
		if strings.ContainsAny(string(char), "=+-*/^()[]{}<>|±×÷∂∆∇∫∑∏√∞≈≠≤≥αβγδϵζηθικλμνξπρστυϕχψω") {
			mathCharCount++
		} else if unicode.IsDigit(char) {
			mathCharCount++
		}
	}

	// If more than 60% of characters are math-related, consider it math
	if totalChars > 0 && float64(mathCharCount)/float64(totalChars) > 0.6 {
		return true
	}

	for _, pattern := range l.mathPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

// isInlineMath mengecek apakah teks adalah matematika inline
func (l *LaTeXConverter) isInlineMath(text string) bool {
	// Simple patterns that suggest inline math
	inlinePatterns := []*regexp.Regexp{
		regexp.MustCompile(`^[a-zA-Z]\s*=\s*.+$`),
		regexp.MustCompile(`^.+\^.+\s*=.+$`),
		regexp.MustCompile(`^[xyz]\s*=\s*\d+$`),
	}

	for _, pattern := range inlinePatterns {
		if pattern.MatchString(text) {
			return true
		}
	}

	return false
}

// processMathExpressions memproses ekspresi matematika
func (l *LaTeXConverter) processMathExpressions(text string) string {
	result := text

	// Replace common math functions
	mathFunctions := map[string]string{
		"sin":  `\sin`,
		"cos":  `\cos`,
		"tan":  `\tan`,
		"log":  `\log`,
		"ln":   `\ln`,
		"lim":  `\lim`,
		"sum":  `\sum`,
		"prod": `\prod`,
		"int":  `\int`,
	}

	for funcName, latexCmd := range mathFunctions {
		pattern := regexp.MustCompile(`\b` + funcName + `\b`)
		result = pattern.ReplaceAllString(result, latexCmd)
	}

	// Handle fractions (basic)
	fractionPattern := regexp.MustCompile(`(\d+)/(\d+)`)
	result = fractionPattern.ReplaceAllString(result, `\frac{$1}{$2}`)

	// Handle exponents
	exponentPattern := regexp.MustCompile(`(\w+)\^(\d+)`)
	result = exponentPattern.ReplaceAllString(result, `$1^{$2}`)

	// Handle square roots
	sqrtPattern := regexp.MustCompile(`sqrt\(([^)]+)\)`)
	result = sqrtPattern.ReplaceAllString(result, `\sqrt{$1}`)

	// Handle integrals
	integralPattern := regexp.MustCompile(`int_(\w+)\^(\w+)\s*(\w+)`)
	result = integralPattern.ReplaceAllString(result, `\int_{$1}^{$2} $3`)

	return result
}

// CreateLatexPreview membuat file LaTeX lengkap untuk preview
func (l *LaTeXConverter) CreateLatexPreview(latexCode, outputPath string) error {
	template := `\documentclass{article}
\usepackage{amsmath}
\usepackage{amssymb}
\usepackage[utf8]{inputenc}
\usepackage{graphicx}
\begin{document}

\title{Hasil Konversi OCR ke LaTeX}
\author{Go LaTeX Converter}
\maketitle

% Hasil konversi dari gambar
` + latexCode + `

\end{document}
`

	return os.WriteFile(outputPath, []byte(template), 0644)
}

// ImageProcessor menangani preprocessing gambar
type ImageProcessor struct{}

// NewImageProcessor membuat instance baru ImageProcessor
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{}
}

// PreprocessImage melakukan preprocessing pada gambar
func (ip *ImageProcessor) PreprocessImage(imagePath string) (string, error) {
	// Untuk kesederhanaan, kita akan langsung menggunakan Tesseract pada gambar asli
	// Dalam implementasi nyata, Anda mungkin ingin menambahkan preprocessing di sini
	return imagePath, nil
}

// SaveProcessedImage menyimpan gambar yang telah diproses
func (ip *ImageProcessor) SaveProcessedImage(img image.Image, outputPath string) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	switch strings.ToLower(filepath.Ext(outputPath)) {
	case ".jpg", ".jpeg":
		return jpeg.Encode(outFile, img, &jpeg.Options{Quality: 90})
	case ".png":
		return png.Encode(outFile, img)
	default:
		return png.Encode(outFile, img)
	}
}

// OCRProcessor menangani ekstraksi teks dari gambar
type OCRProcessor struct {
	tesseractPath string
}

// NewOCRProcessor membuat instance baru OCRProcessor
func NewOCRProcessor() *OCRProcessor {
	processor := &OCRProcessor{}

	// Cari path Tesseract
	if path, err := processor.findTesseract(); err == nil {
		processor.tesseractPath = path
	} else {
		log.Printf("Peringatan: Tesseract tidak ditemukan: %v", err)
	}

	return processor
}

// findTesseract mencari lokasi Tesseract di sistem
func (ocr *OCRProcessor) findTesseract() (string, error) {
	// Coba lokasi umum
	possiblePaths := []string{
		"tesseract",
		"/usr/bin/tesseract",
		"/usr/local/bin/tesseract",
		"/opt/homebrew/bin/tesseract",
		"C:\\Program Files\\Tesseract-OCR\\tesseract.exe",
	}

	for _, path := range possiblePaths {
		if _, err := exec.LookPath(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("tesseract tidak ditemukan di sistem")
}

// ExtractText mengekstrak teks dari gambar menggunakan Tesseract OCR
func (ocr *OCRProcessor) ExtractText(imagePath string) (string, error) {
	if ocr.tesseractPath == "" {
		return "", fmt.Errorf("tesseract tidak tersedia")
	}

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command(ocr.tesseractPath, imagePath, "stdout", "-l", "eng", "--psm", "6")
	} else {
		cmd = exec.Command(ocr.tesseractPath, imagePath, "stdout", "-l", "eng", "--psm", "6")
	}

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("tesseract error: %v, %s", err, stderr.String())
	}

	return strings.TrimSpace(out.String()), nil
}

// ExtractTextWithLanguage mengekstrak teks dengan bahasa tertentu
func (ocr *OCRProcessor) ExtractTextWithLanguage(imagePath, language string) (string, error) {
	if ocr.tesseractPath == "" {
		return "", fmt.Errorf("tesseract tidak tersedia")
	}

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command(ocr.tesseractPath, imagePath, "stdout", "-l", language, "--psm", "6")
	} else {
		cmd = exec.Command(ocr.tesseractPath, imagePath, "stdout", "-l", language, "--psm", "6")
	}

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("tesseract error: %v, %s", err, stderr.String())
	}

	return strings.TrimSpace(out.String()), nil
}

// SimpleTextExtractor ekstraktor teks sederhana tanpa Tesseract (untuk testing)
type SimpleTextExtractor struct{}

// NewSimpleTextExtractor membuat instance baru SimpleTextExtractor
func NewSimpleTextExtractor() *SimpleTextExtractor {
	return &SimpleTextExtractor{}
}

// ExtractText mengekstrak teks dari gambar (dummy implementation)
func (ste *SimpleTextExtractor) ExtractText(imagePath string) (string, error) {
	// Ini adalah implementasi dummy untuk testing
	// Dalam implementasi nyata, Anda akan menggunakan OCR
	log.Printf("Membaca gambar: %s", imagePath)

	// Untuk demo, kita return teks contoh
	return "E = mc^2\n\n∫ from 0 to 1 x^2 dx = 1/3\n\nlim x→∞ (1 + 1/x)^x = e", nil
}

// createSampleImage membuat gambar contoh untuk testing
func createSampleImage(outputPath string) error {
	width, height := 800, 600
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Background putih
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Tambahkan beberapa teks matematika
	texts := []string{
		"E = mc^2",
		"∫₀¹ x² dx = 1/3",
		"lim x→∞ (1 + 1/x)ˣ = e",
		"∂²u/∂t² = c² ∇²u",
	}

	point := fixed.P(50, 50)
	for _, text := range texts {
		point = addLabel(img, point.X, point.Y, text)
		point.Y += 80
	}

	// Simpan gambar
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return png.Encode(outFile, img)
}

// addLabel menambahkan teks ke gambar
func addLabel(img *image.RGBA, x, y fixed.Int26_6, label string) fixed.Point26_6 {
	col := color.Black
	point := fixed.P(x.Round(), y.Round())

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)

	return d.Dot
}

// fileExists mengecek apakah file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// main function
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Go LaTeX OCR Converter")
		fmt.Println("Usage: go run main.go <image_path> [language]")
		fmt.Println("Example: go run main.go equation.png eng")
		fmt.Println("Example: go run main.go equation.png ind")
		fmt.Println("\nUntuk membuat gambar contoh:")
		fmt.Println("go run main.go --create-sample")
		os.Exit(1)
	}

	if os.Args[1] == "--create-sample" {
		fmt.Println("Membuat gambar contoh...")
		err := createSampleImage("sample_equation.png")
		if err != nil {
			log.Fatalf("Gagal membuat gambar contoh: %v", err)
		}
		fmt.Println("Gambar contoh dibuat: sample_equation.png")
		return
	}

	imagePath := os.Args[1]
	language := "eng"
	if len(os.Args) > 2 {
		language = os.Args[2]
	}

	// Validasi file
	if !fileExists(imagePath) {
		log.Fatalf("File tidak ditemukan: %s", imagePath)
	}

	fmt.Printf("Memproses gambar: %s\n", imagePath)
	fmt.Printf("Bahasa OCR: %s\n", language)

	// Inisialisasi processor
	imageProcessor := NewImageProcessor()
	ocrProcessor := NewOCRProcessor()
	latexConverter := NewLaTeXConverter()

	// Preprocess gambar
	processedImagePath, err := imageProcessor.PreprocessImage(imagePath)
	if err != nil {
		log.Fatalf("Error preprocessing gambar: %v", err)
	}

	// Ekstrak teks dengan OCR
	var extractedText string

	if ocrProcessor.tesseractPath != "" {
		fmt.Println("Menggunakan Tesseract OCR...")
		if language == "eng" {
			extractedText, err = ocrProcessor.ExtractText(processedImagePath)
		} else {
			extractedText, err = ocrProcessor.ExtractTextWithLanguage(processedImagePath, language)
		}

		if err != nil {
			log.Printf("Peringatan OCR: %v", err)
			log.Println("Menggunakan ekstraktor sederhana...")
			simpleExtractor := NewSimpleTextExtractor()
			extractedText, err = simpleExtractor.ExtractText(processedImagePath)
		}
	} else {
		fmt.Println("Tesseract tidak ditemukan, menggunakan ekstraktor sederhana...")
		simpleExtractor := NewSimpleTextExtractor()
		extractedText, err = simpleExtractor.ExtractText(processedImagePath)
	}

	if err != nil {
		log.Fatalf("Error ekstraksi teks: %v", err)
	}

	fmt.Printf("\nTeks terdeteksi:\n%s\n", extractedText)

	// Konversi ke LaTeX
	latexCode := latexConverter.ConvertToLatex(extractedText)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("HASIL KONVERSI LaTeX:")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(latexCode)

	// Simpan hasil
	outputFile := "output.tex"
	err = os.WriteFile(outputFile, []byte(latexCode), 0644)
	if err != nil {
		log.Fatalf("Error menyimpan file: %v", err)
	}

	fmt.Printf("\nHasil disimpan ke: %s\n", outputFile)

	// Buat preview LaTeX lengkap
	previewFile := "preview.tex"
	err = latexConverter.CreateLatexPreview(latexCode, previewFile)
	if err != nil {
		log.Fatalf("Error membuat preview: %v", err)
	}

	fmt.Printf("Preview LaTeX lengkap dibuat: %s\n", previewFile)

	// Informasi kompilasi
	fmt.Println("\nUntuk mengkompilasi file LaTeX:")
	fmt.Printf("pdflatex %s\n", previewFile)
}
