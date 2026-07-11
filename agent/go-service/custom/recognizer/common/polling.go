package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

const minimumPollingInterval = 200 * time.Millisecond

type CommonPollingRecognizerRunner struct{}

type CommonPollingRecognitionParam struct {
	Timeout  int64               `json:"timeout"`
	Interval int64               `json:"interval"`
	Type     maa.RecognitionType `json:"type"`
	Param    json.RawMessage     `json:"param"`
}

var _ maa.CustomRecognitionRunner = &CommonPollingRecognizerRunner{}

func (r *CommonPollingRecognizerRunner) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	var config CommonPollingRecognitionParam
	if err := decodeStrictJSON([]byte(arg.CustomRecognitionParam), &config); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal CommonPollingRecognition parameters")
		return nil, false
	}

	timeout := time.Duration(config.Timeout) * time.Millisecond
	interval := time.Duration(config.Interval) * time.Millisecond
	if timeout <= 0 {
		log.Error().Int64("timeout", config.Timeout).Msg("CommonPollingRecognition timeout must be positive")
		return nil, false
	}
	if interval < minimumPollingInterval {
		log.Error().Int64("interval", config.Interval).Msg("CommonPollingRecognition interval must be at least 200ms")
		return nil, false
	}

	recognitionParam, err := bindRecognitionParam(config.Type, config.Param)
	if err != nil {
		log.Error().Err(err).Str("type", string(config.Type)).Msg("invalid CommonPollingRecognition recognition parameters")
		return nil, false
	}

	deadline := time.Now().Add(timeout)
	controller := ctx.GetTasker().GetController()
	for {
		if !controller.PostScreencap().Wait().Success() {
			log.Warn().Msg("CommonPollingRecognition failed to capture a fresh frame")
		} else if img, err := controller.CacheImage(); err != nil {
			log.Warn().Err(err).Msg("CommonPollingRecognition failed to get the captured frame")
		} else {
			detail, err := ctx.RunRecognitionDirect(config.Type, recognitionParam, img)
			if err != nil {
				log.Warn().Err(err).Str("type", string(config.Type)).Msg("CommonPollingRecognition attempt failed")
			} else if detail.Hit && !time.Now().After(deadline) {
				return &maa.CustomRecognitionResult{
					Box:    detail.Box,
					Detail: detail.DetailJson,
				}, true
			}
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, false
		}
		if remaining < interval {
			time.Sleep(remaining)
			return nil, false
		}
		time.Sleep(interval)
	}
}

func bindRecognitionParam(recognitionType maa.RecognitionType, raw json.RawMessage) (maa.RecognitionParam, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		raw = json.RawMessage(`{}`)
	}

	var param maa.RecognitionParam
	switch recognitionType {
	case maa.RecognitionTypeDirectHit:
		param = &maa.DirectHitParam{}
	case maa.RecognitionTypeTemplateMatch:
		param = &maa.TemplateMatchParam{}
	case maa.RecognitionTypeFeatureMatch:
		param = &maa.FeatureMatchParam{}
	case maa.RecognitionTypeColorMatch:
		param = &maa.ColorMatchParam{}
	case maa.RecognitionTypeOCR:
		param = &maa.OCRParam{}
	case maa.RecognitionTypeNeuralNetworkClassify:
		param = &maa.NeuralNetworkClassifyParam{}
	case maa.RecognitionTypeNeuralNetworkDetect:
		param = &maa.NeuralNetworkDetectParam{}
	case maa.RecognitionTypeAnd:
		param = &maa.AndRecognitionParam{}
	case maa.RecognitionTypeOr:
		param = &maa.OrRecognitionParam{}
	case maa.RecognitionTypeCustom:
		return nil, errors.New("nested Custom recognition is not supported")
	default:
		return nil, fmt.Errorf("unsupported recognition type %q", recognitionType)
	}

	if err := decodeStrictJSON(raw, param); err != nil {
		return nil, fmt.Errorf("parameters do not match recognition type %q: %w", recognitionType, err)
	}
	return param, nil
}

func decodeStrictJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values are not allowed")
		}
		return err
	}
	return nil
}
