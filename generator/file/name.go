package file

import (
	"math/rand"
	"strings"
	"time"
)

// Configuration: Set to true for romanized Chinese, Thai, Japanese, and Arabic, false for native scripts
const useRomanized = false // Affects Chinese, Thai, Japanese, Arabic; Korean always uses Hangul

// Generate an English-like word (Latin-based)
func generateEnglishWord() string {
	vowels := []string{"a", "e", "i", "o", "u", "ai", "ea", "ou"}
	consonants := []string{"b", "c", "d", "f", "g", "h", "j", "k", "l", "m", "n", "p", "r", "s", "t", "v", "w", "y", "sh", "ch", "th"}
	length := 3 + rand.Intn(5) // 3 to 7 letters

	word := ""
	for i := 0; i < length; i++ {
		if i%2 == 0 {
			word += consonants[rand.Intn(len(consonants))]
		} else {
			word += vowels[rand.Intn(len(vowels))]
		}
	}
	return word
}

// Generate a non-English word based on specified language
func generateNonEnglishWord(lang string) string {
	langPatterns := map[string]struct {
		native        []string // Native script characters or syllables
		romanized     []string // Romanized equivalents (used for Chinese, Thai, Japanese, Arabic if useRomanized = true)
		suffixes      []string // Optional suffixes (native or romanized)
		syllableCount int      // Number of syllables (1-2)
	}{
		"vi": { // Vietnamese-like (e.g., nâm, hỏa, with accents)
			native:        []string{"nâm", "hỏa", "lân", "thư", "mình", "ngọc", "tâm", "việt", "phố", "sông", "hà", "nội", "đà", "nẵng", "huế", "cần"},
			romanized:     []string{"nam", "hoa", "lan", "thu", "minh", "ngoc", "tam", "viet", "pho", "song", "ha", "noi", "da", "nang", "hue", "can"},
			suffixes:      []string{"", "", ""}, // No suffixes
			syllableCount: 1 + rand.Intn(2),
		},
		"ge": { // German-like (e.g., mü, schön, with umlauts and ß)
			native:        []string{"mü", "schön", "wald", "stern", "bau", "feld", "himmel", "licht", "tag", "nacht", "straße", "berg", "fluss", "baum", "grün", "weiß"},
			romanized:     []string{"mue", "schoen", "wald", "stern", "bau", "feld", "himmel", "licht", "tag", "nacht", "strasse", "berg", "fluss", "baum", "gruen", "weiss"},
			suffixes:      []string{"en", "er", "d", "e", "in"},
			syllableCount: 1 + rand.Intn(2),
		},
		"cn": { // Chinese-like (e.g., 好, 星, Pinyin: hao, xing)
			native:        []string{"好", "星", "美", "兰", "君", "伟", "青", "书", "天", "花", "月", "山", "水", "风", "云", "龙", "凤", "春", "秋"},
			romanized:     []string{"hao", "xing", "mei", "lan", "jun", "wei", "qing", "shu", "tian", "hua", "yue", "shan", "shui", "feng", "yun", "long", "feng", "chun", "qiu"},
			suffixes:      []string{"", "", ""}, // No suffixes
			syllableCount: 1 + rand.Intn(2),
		},
		"kn": { // Korean-like (e.g., 하, 나, 별, always native Hangul)
			native:        []string{"하", "나", "별", "미", "지", "라", "고", "타", "영", "수", "강", "산", "바", "람", "꽃", "하늘", "달", "빛", "소리"},
			romanized:     []string{"하", "나", "별", "미", "지", "라", "고", "타", "영", "수", "강", "산", "바", "람", "꽃", "하늘", "달", "빛", "소리"}, // Ignored, always native
			suffixes:      []string{"", "ㄴ", "ㅁ", "이"},                                                                               // Native Hangul suffixes
			syllableCount: 1 + rand.Intn(2),
		},
		"tl": { // Thai-like (e.g., ชัย, สุข, Romanized: chai, suk)
			native:        []string{"ชัย", "สุข", "รถ", "ผัด", "ใหม่", "น้ำ", "ขาว", "ลม", "ดิน", "ไฟ", "ฟ้า", "ต้น", "ใบ", "หิน", "แสง", "เงา"},
			romanized:     []string{"chai", "suk", "rot", "phat", "mai", "nam", "khao", "lom", "din", "fai", "fa", "ton", "bai", "hin", "saeng", "ngao"},
			suffixes:      []string{"", "ต", "น", "ม"}, // Native suffixes (romanized: t, n, m)
			syllableCount: 1 + rand.Intn(2),
		},
		"jp": { // Japanese-like (e.g., さ, く, Hiragana, Romanized: sa, ku)
			native:        []string{"さ", "く", "ら", "み", "な", "き", "ゆ", "め", "ひ", "ろ", "か", "ぜ", "そ", "ら", "つ", "き", "や", "ま", "は", "な"},
			romanized:     []string{"sa", "ku", "ra", "mi", "na", "ki", "yu", "me", "hi", "ro", "ka", "ze", "so", "ra", "tsu", "ki", "ya", "ma", "ha", "na"},
			suffixes:      []string{"", "ん", "い", "う"}, // Native suffixes (romanized: n, i, u)
			syllableCount: 1 + rand.Intn(2),
		},
		"ar": { // Arabic-like (e.g., نور, سلا, Transliterated: nur, sala)
			native:        []string{"نور", "سلا", "رح", "مح", "زي", "حل", "جم", "فر", "قمر", "شمس", "نجم", "سماء", "بحر", "رمل", "ضوء", "هواء"},
			romanized:     []string{"nur", "sala", "rah", "mah", "zi", "hal", "jam", "far", "qamar", "shams", "najm", "sama", "bahr", "raml", "daw", "hawa"},
			suffixes:      []string{"", "ة", "ي", "ات"}, // Native suffixes (romanized: a, i, at)
			syllableCount: 1 + rand.Intn(2),
		},
	}

	pattern, exists := langPatterns[lang]
	if !exists {
		// Fallback to Vietnamese if language not found
		pattern = langPatterns["vi"]
	}

	word := ""
	// Generate syllables
	for i := 0; i < pattern.syllableCount; i++ {
		if lang == "kn" || !useRomanized {
			word += pattern.native[rand.Intn(len(pattern.native))]
		} else {
			word += pattern.romanized[rand.Intn(len(pattern.romanized))]
		}
	}

	// Add suffix with 50% probability
	if rand.Float32() < 0.5 && len(pattern.suffixes) > 0 && pattern.suffixes[0] != "" {
		if lang == "kn" || !useRomanized {
			word += pattern.suffixes[rand.Intn(len(pattern.suffixes))]
		} else {
			word += pattern.suffixes[rand.Intn(len(pattern.suffixes))]
		}
	}

	return word
}

// Check if a word is already in the selected list
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Generate a filename with 2 to 5 words, all from one randomly chosen language
func generateFilename(langs []string) string {
	rand.Seed(time.Now().UnixNano())

	// Randomly choose one language from the provided list
	lang := langs[rand.Intn(len(langs))]

	// Randomly choose number of words (2 to 5)
	numWords := 2 + rand.Intn(4)

	// Collect words
	var selectedWords []string

	// Generate words in the chosen language
	for i := 0; i < numWords; i++ {
		var word string
		if lang == "en" {
			word = generateEnglishWord()
		} else {
			word = generateNonEnglishWord(lang)
		}
		// Avoid duplicates
		for contains(selectedWords, word) {
			if lang == "en" {
				word = generateEnglishWord()
			} else {
				word = generateNonEnglishWord(lang)
			}
		}
		selectedWords = append(selectedWords, word)
	}

	// Join words with hyphens
	return strings.Join(selectedWords, " ")
}
