package marian

type Config struct {
	VocabSize           int      `json:"vocab_size"`
	DecoderVocabSize    int      `json:"decoder_vocab_size"`
	EosTokenID          int64    `json:"eos_token_id"`
	BosTokenID          int64    `json:"bos_token_id"`
	PadTokenID          int64    `json:"pad_token_id"`
	DecoderStartTokenID int64    `json:"decoder_start_token_id"`
	MaxLength           int      `json:"max_length"`
	ModelMaxLength      int      `json:"model_max_length"`
	BadWordsIDs         [][]int  `json:"bad_words_ids"`
}

func (t *Config) NormalizeConfig() {
	if t.DecoderVocabSize == 0 {
		t.DecoderVocabSize = t.VocabSize
	}
	if t.ModelMaxLength == 0 {
		switch {
		case t.MaxLength > 0:
			t.ModelMaxLength = t.MaxLength
		default:
			t.ModelMaxLength = 512
		}
	}
	if t.BosTokenID == 0 {
		t.BosTokenID = t.EosTokenID
	}
}