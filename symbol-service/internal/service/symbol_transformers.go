package service

import (
	"encoding/base64"
	"encoding/json"
	v1 "symbol-service/api/symbol/v1"
	"symbol-service/internal/biz"
)

func toBizSymbol(s *v1.Symbol) *biz.Symbol {
	return &biz.Symbol{
		Id:              s.Id,
		Project:         s.ProjectId,
		Uid:             s.Uid,
		Label:           s.Label,
		ClassName:       s.ClassName,
		ComponentTarget: s.ComponentTarget,
		Version:         s.Version,
		Data: &biz.SymbolData{
			Project: s.ProjectId,
			Data:    s.Data,
		},
	}
}

func toBizSymbolFromRequest(s *v1.CreateSymbolRequest) *biz.Symbol {
	return &biz.Symbol{
		Project:         s.ProjectId,
		Uid:             s.Uid,
		Label:           s.Label,
		ClassName:       s.ClassName,
		ComponentTarget: s.ComponentTarget,
		Version:         s.Version,
		Data: &biz.SymbolData{
			Project: s.ProjectId,
			Data:    s.Data,
		},
	}
}

func toV1Symbol(s *biz.Symbol) *v1.Symbol {
	var data []byte
	if s.Data != nil {
		data = s.Data.Data
	}
	return &v1.Symbol{
		Id:              s.Id,
		ProjectId:       s.Project,
		Uid:             s.Uid,
		Label:           s.Label,
		ClassName:       s.ClassName,
		ComponentTarget: s.ComponentTarget,
		Version:         s.Version,
		Data:            data,
	}
}

func toBizListSymbolsOptions(in *v1.ListSymbolsRequest) *biz.ListSymbolsOptions {

	options := &biz.ListSymbolsOptions{
		ProjectID: in.ProjectId,
		PageSize:  in.PageSize,
	}

	if cursorStr, err := base64.StdEncoding.DecodeString(in.PageToken); err != nil {
		options.Cursor = fromV1PageTokenToSymbolCursor(cursorStr)
	}

	return options
}

// toV1PagetToken encodes a SymbolCursor to a JSON string for pagination.
func toV1PageToken(cursor *biz.SymbolCursor) string {
	if bytes, err := json.Marshal(cursor); err == nil {
		return base64.StdEncoding.EncodeToString(bytes)
	} else {
		return ""
	}
}

func fromV1PageTokenToSymbolCursor(pageToken []byte) *biz.SymbolCursor {
	symbolCursor := biz.SymbolCursor{}
	if err := json.Unmarshal(pageToken, &symbolCursor); err == nil {
		return &symbolCursor
	}

	return &symbolCursor
}
