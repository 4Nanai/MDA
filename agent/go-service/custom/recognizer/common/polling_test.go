package common

import (
	"encoding/json"
	"testing"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
)

func TestBindRecognitionParam(t *testing.T) {
	t.Run("matching type and parameters", func(t *testing.T) {
		param, err := bindRecognitionParam(maa.RecognitionTypeTemplateMatch, json.RawMessage(`{
            "roi": [1250, 112, 30, 107],
            "template": ["Common/RedDot.png"]
        }`))
		if err != nil {
			t.Fatalf("bindRecognitionParam() error = %v", err)
		}
		if _, ok := param.(*maa.TemplateMatchParam); !ok {
			t.Fatalf("bindRecognitionParam() type = %T, want *maa.TemplateMatchParam", param)
		}
	})

	t.Run("mismatched type and parameters", func(t *testing.T) {
		_, err := bindRecognitionParam(maa.RecognitionTypeOCR, json.RawMessage(`{
            "template": ["Common/RedDot.png"]
        }`))
		if err == nil {
			t.Fatal("bindRecognitionParam() error = nil, want mismatched parameter error")
		}
	})

	t.Run("nested custom recognition", func(t *testing.T) {
		_, err := bindRecognitionParam(maa.RecognitionTypeCustom, json.RawMessage(`{}`))
		if err == nil {
			t.Fatal("bindRecognitionParam() error = nil, want unsupported type error")
		}
	})
}
