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

type Tokenizer interface {
	// Encode encodes a single source sentence into token IDs.
	// If addEOS is true, EOS token is appended.
	Encode(text string, addEOS bool) ([]int64, error)

	// EncodeBatch encodes a batch of sentences and returns:
	//  - inputIDs: shape (batch, maxLen)
	//  - attentionMask: shape (batch, maxLen) with 1 for tokens and 0 for padding.
	EncodeBatch(texts []string) (inputIDs [][]int64, attentionMask [][]int64, err error)

	// Decode converts token IDs back to a target sentence.
	// If skipSpecial is true, EOS / PAD / UNK are removed before decoding.
	Decode(ids []int64, skipSpecial bool) (string, error)

	// Close releases any underlying native resources (SentencePiece, Marian, etc.).
	Close()
}
