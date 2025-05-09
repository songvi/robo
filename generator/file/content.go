package file

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/unidoc/unioffice/document"
	"github.com/xuri/excelize/v2"
)

// FileContentGenerator generates file content based on extension and size
type FileContentGenerator struct {
	RepositoryPath string // Base directory for storing files
}

// NewFileContentGenerator initializes a new FileContentGenerator
func NewFileContentGenerator(repositoryPath string) *FileContentGenerator {
	return &FileContentGenerator{
		RepositoryPath: repositoryPath,
	}
}

// Generate a rich sentence in the specified language, defaulting to English
func generateSentence(lang string) string {
	type sentencePattern struct {
		subjects   []string
		verbs      []string
		objects    []string
		adjectives []string
		adverbs    []string
		connectors []string
		isSOV      bool // Subject-Object-Verb (e.g., Korean, Japanese) vs Subject-Verb-Object
	}

	patterns := map[string]sentencePattern{
		"en": {
			subjects:   []string{"The sky", "The forest", "A bird", "The river", "The moon", "A child", "The wind", "The mountain"},
			verbs:      []string{"sings", "dances", "flows", "shines", "whispers", "climbs", "soars", "rests"},
			objects:    []string{"a sweet melody", "through the trees", "gently", "in the night", "with grace", "to the stars", "peacefully", "under the sun"},
			adjectives: []string{"tranquil", "radiant", "serene", "majestic", "gentle", "vibrant", "quiet", "sparkling"},
			adverbs:    []string{"beautifully", "gracefully", "silently", "boldly", "softly", "swiftly", "calmly", "elegantly"},
			connectors: []string{"and", "while", "as", "under", "beneath", "across", "within", "beyond"},
			isSOV:      false,
		},
		"cn": {
			subjects:   []string{"天空", "森林", "鸟儿", "河流", "月亮", "孩子", "风", "山峰"},
			verbs:      []string{"歌唱", "舞动", "流动", "闪耀", "低语", "攀登", "翱翔", "休息"},
			objects:    []string{"甜美的旋律", "穿过树林", "轻轻地", "在夜晚", "优雅地", "向星星", "平静地", "在阳光下"},
			adjectives: []string{"宁静的", "光芒四射的", "平静的", "雄伟的", "温柔的", "生动的", "安静的", "闪亮的"},
			adverbs:    []string{"美丽地", "优雅地", "静静地", "大胆地", "柔和地", "迅速地", "平静地", "高雅地"},
			connectors: []string{"并且", "当", "如同", "在…之下", "在…下面", "穿过", "在…之中", "超越"},
			isSOV:      false,
		},
		"kn": {
			subjects:   []string{"하늘", "숲", "새", "강", "달", "아이", "바람", "산"},
			verbs:      []string{"노래한다", "춤춘다", "흐른다", "빛난다", "속삭인다", "오른다", "날아오른다", "쉰다"},
			objects:    []string{"달콤한 멜로디를", "나무 사이를", "부드럽게", "밤에", "우아하게", "별을 향해", "평화롭게", "태양 아래"},
			adjectives: []string{"고요한", "찬란한", "평온한", "웅장한", "온화한", "생생한", "조용한", "반짝이는"},
			adverbs:    []string{"아름답게", "우아하게", "조용히", "대담하게", "부드럽게", "빠르게", "차분히", "고상하게"},
			connectors: []string{"그리고", "하면서", "처럼", "아래", "밑에", "건너", "안에", "넘어"},
			isSOV:      true,
		},
		"tl": {
			subjects:   []string{"Ang langit", "Ang gubat", "Isang ibon", "Ang ilog", "Ang buwan", "Isang bata", "Ang hangin", "Ang bundok"},
			verbs:      []string{"kumakanta", "sumasayaw", "umaagos", "kumikinang", "bumubulong", "umaakyat", "lumilipad", "nagpapahinga"},
			objects:    []string{"isang matamis na himig", "sa mga puno", "nang marahan", "sa gabi", "nang may biyaya", "patungo sa mga bituin", "nang mapayapa", "sa ilalim ng araw"},
			adjectives: []string{"tahimik", "maningning", "payapa", "marilag", "banayad", "buhay na buhay", "katahimikan", "kumikislap"},
			adverbs:    []string{"nang maganda", "nang magaan", "nang tahimik", "nang matapang", "nang malambot", "nang mabilis", "nang kalmado", "nang elegante"},
			connectors: []string{"at", "habang", "tulad ng", "sa ilalim", "sa baba", "sa kabila", "sa loob", "lampas sa"},
			isSOV:      false,
		},
		"jp": {
			subjects:   []string{"空", "森", "鳥", "川", "月", "子", "風", "山"},
			verbs:      []string{"歌う", "踊る", "流れる", "輝く", "囁く", "登る", "飛ぶ", "休む"},
			objects:    []string{"甘いメロディーを", "木々の間を", "優しく", "夜に", "優雅に", "星に向かって", "平和に", "太陽の下で"},
			adjectives: []string{"静かな", "輝く", "穏やかな", "壮大な", "優しい", "鮮やかな", "静寂な", "きらめく"},
			adverbs:    []string{"美しく", "優雅に", "静かに", "大胆に", "柔らかく", "速く", "穏やかに", "上品に"},
			connectors: []string{"そして", "ながら", "ように", "下で", "下に", "越えて", "中に", "超えて"},
			isSOV:      true,
		},
		"ar": {
			subjects:   []string{"السماء", "الغابة", "طائر", "النهر", "القمر", "طفل", "الريح", "الجبل"},
			verbs:      []string{"يغني", "يرقص", "يتدفق", "يتألق", "يهمس", "يتسلق", "يحلق", "يرتاح"},
			objects:    []string{"لحناً عذباً", "بين الأشجار", "بلطف", "في الليل", "بأناقة", "نحو النجوم", "بسلام", "تحت الشمس"},
			adjectives: []string{"هادئة", "مشرقة", "ساكنة", "مهيبة", "لطيفة", "نابضة بالحياة", "صامتة", "متلألئة"},
			adverbs:    []string{"بجمال", "بأناقة", "بهدوء", "بجرأة", "بلطف", "بسرعة", "بهدوء", "بأناقة"},
			connectors: []string{"و", "بينما", "كما", "تحت", "أسفل", "عبر", "داخل", "وراء"},
			isSOV:      false,
		},
	}

	pattern, exists := patterns[lang]
	if !exists {
		pattern = patterns["en"] // Default to English
	}

	// Randomly decide sentence complexity
	hasAdjective := rand.Float32() < 0.7
	hasAdverb := rand.Float32() < 0.6
	hasConnector := rand.Float32() < 0.4

	// First clause
	subject := pattern.subjects[rand.Intn(len(pattern.subjects))]
	verb := pattern.verbs[rand.Intn(len(pattern.verbs))]
	object := pattern.objects[rand.Intn(len(pattern.objects))]
	var adjective, adverb string
	if hasAdjective {
		adjective = pattern.adjectives[rand.Intn(len(pattern.adjectives))] + " "
	}
	if hasAdverb {
		adverb = pattern.adverbs[rand.Intn(len(pattern.adverbs))]
	}

	var sentence string
	if pattern.isSOV {
		sentence = fmt.Sprintf("%s%s %s %s", adjective, subject, object, verb)
		if hasAdverb {
			sentence += " " + adverb
		}
	} else {
		sentence = fmt.Sprintf("%s%s %s %s", adjective, subject, verb, object)
		if hasAdverb {
			sentence += " " + adverb
		}
	}

	// Add a second clause with connector
	if hasConnector {
		connector := pattern.connectors[rand.Intn(len(pattern.connectors))]
		subject2 := pattern.subjects[rand.Intn(len(pattern.subjects))]
		verb2 := pattern.verbs[rand.Intn(len(pattern.verbs))]
		object2 := pattern.objects[rand.Intn(len(pattern.objects))]
		hasAdjective2 := rand.Float32() < 0.5
		var adjective2 string
		if hasAdjective2 {
			adjective2 = pattern.adjectives[rand.Intn(len(pattern.adjectives))] + " "
		}

		var clause string
		if pattern.isSOV {
			clause = fmt.Sprintf("%s%s %s %s", adjective2, subject2, object2, verb2)
		} else {
			clause = fmt.Sprintf("%s%s %s %s", adjective2, subject2, verb2, object2)
		}
		sentence += fmt.Sprintf(" %s %s", connector, clause)
	}

	// Add punctuation
	if lang == "cn" || lang == "jp" {
		sentence += "。"
	} else if lang == "ar" {
		sentence += "."
	} else {
		sentence += "."
	}

	return sentence
}

// GenerateContent generates file content and saves it to the repository
func (g *FileContentGenerator) GenerateContent(file *File, lang string) error {
	rand.Seed(time.Now().UnixNano())

	// Create the full file path in the repository
	fullPath := filepath.Join(g.RepositoryPath, file.FilePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	switch strings.ToLower(file.FileExtension) {
	case "txt":
		// Generate text content
		var content strings.Builder
		targetSize := file.FileSize
		if targetSize < 1024 {
			targetSize = 1024 // Minimum 1KB
		}
		if targetSize > 5*1024*1024 {
			targetSize = 5 * 1024 * 1024 // Max 5MB
		}

		for content.Len() < targetSize {
			content.WriteString(generateSentence(lang) + "\n")
		}
		// Truncate to exact size
		contentStr := content.String()
		if len(contentStr) > targetSize {
			contentStr = contentStr[:targetSize]
		}

		if err := os.WriteFile(fullPath, []byte(contentStr), 0644); err != nil {
			return fmt.Errorf("failed to write txt file: %v", err)
		}
		file.FileContent = "Generated text content"

	case "pdf":
		// Generate PDF with non-Latin text
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddUTF8Font("NotoSans", "", "NotoSans-Regular.ttf") // Assumes NotoSans-Regular.ttf in working directory
		pdf.AddPage()
		pdf.SetFont("NotoSans", "", 12)
		targetSize := file.FileSize
		if targetSize < 1024 {
			targetSize = 1024
		}
		if targetSize > 5*1024*1024 {
			targetSize = 5 * 1024 * 1024
		}

		for i := 0; pdf.GetY() < 270 && i*len(generateSentence(lang)) < targetSize; i++ {
			pdf.Write(5, generateSentence(lang)+"\n")
		}

		if err := pdf.OutputFileAndClose(fullPath); err != nil {
			return fmt.Errorf("failed to write pdf file: %v", err)
		}
		file.FileContent = "Generated PDF content"

	case "docx":
		// Generate DOCX with non-Latin text
		doc := document.New()
		targetSize := file.FileSize
		if targetSize < 1024 {
			targetSize = 1024
		}
		if targetSize > 5*1024*1024 {
			targetSize = 5 * 1024 * 1024
		}

		for i := 0; i*len(generateSentence(lang)) < targetSize; i++ {
			para := doc.AddParagraph()
			para.AddRun().AddText(generateSentence(lang))
		}

		if err := doc.SaveToFile(fullPath); err != nil {
			return fmt.Errorf("failed to write docx file: %v", err)
		}
		file.FileContent = "Generated DOCX content"

	case "xlsx":
		// Generate XLSX with non-Latin text
		f := excelize.NewFile()
		targetSize := file.FileSize
		if targetSize < 1024 {
			targetSize = 1024
		}
		if targetSize > 5*1024*1024 {
			targetSize = 5 * 1024 * 1024
		}

		for i := 1; i <= 100 && i*len(generateSentence(lang)) < targetSize; i++ {
			cell := fmt.Sprintf("A%d", i)
			f.SetCellValue("Sheet1", cell, generateSentence(lang))
		}

		if err := f.SaveAs(fullPath); err != nil {
			return fmt.Errorf("failed to write xlsx file: %v", err)
		}
		file.FileContent = "Generated XLSX content"

	case "jpeg", "png":
		// Generate a simple image
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{255, 0, 0, 255}}, image.Point{}, draw.Src)

		targetSize := file.FileSize
		if targetSize < 1024 {
			targetSize = 1024
		}
		if targetSize > 5*1024*1024 {
			targetSize = 5 * 1024 * 1024
		}

		f, err := os.Create(fullPath)
		if err != nil {
			return fmt.Errorf("failed to create image file: %v", err)
		}
		defer f.Close()

		if file.FileExtension == "jpeg" {
			if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 75}); err != nil {
				return fmt.Errorf("failed to write jpeg file: %v", err)
			}
		} else {
			if err := png.Encode(f, img); err != nil {
				return fmt.Errorf("failed to write png file: %v", err)
			}
		}

		// Pad file to reach target size
		f.Seek(0, 2)
		currentSize, _ := f.Seek(0, 1)
		if int(currentSize) < targetSize {
			padding := make([]byte, targetSize-int(currentSize))
			rand.Read(padding)
			f.Write(padding)
		}
		file.FileContent = "Generated image content"

	case "bin":
		// Generate binary content
		targetSize := file.FileSize
		if targetSize < 1024*1024 {
			targetSize = 1024 * 1024 // Minimum 1MB
		}
		if targetSize > 1024*1024*1024 {
			targetSize = 1024 * 1024 * 1024 // Max 1GB
		}

		data := make([]byte, targetSize)
		rand.Read(data) // Non-null random bytes

		if err := os.WriteFile(fullPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write bin file: %v", err)
		}
		file.FileContent = "Generated binary content"

	default:
		return fmt.Errorf("unsupported file extension: %s", file.FileExtension)
	}

	return nil
}

// func main() {
// 	// Initialize generator
// 	generator := NewFileContentGenerator("./generated_files")

// 	// Example file
// 	file := &File{
// 		Name:          "하-별-꽃-영",
// 		Description:   "Generated file in kn",
// 		FileExtension: "docx",
// 		FileSize:      1024 * 1024, // 1MB
// 		FilePath:      "/강산/하-별-꽃-영.docx",
// 	}

// 	// Generate content
// 	if err := generator.GenerateContent(file, "kn"); err != nil {
// 		fmt.Printf("Error generating content: %v\n", err)
// 		return
// 	}

// 	// Print file info
// 	fileJSON, _ := json.MarshalIndent(file, "", "  ")
// 	fmt.Printf("Generated file: %s\n", string(fileJSON))
// }
