package marian

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

	// Config returns the tokenizer configuration.
	//
	// The configuration is loaded and cached during tokenizer initialization and
	// remains immutable for the lifetime of the tokenizer. Implementations may
	// cache the parsed configuration and return it on subsequent calls.
	//
	// The returned pointer refers to an internal cached copy and must not be
	// modified by the caller.
	Config() (*Config, error)

	// Close releases any underlying native resources (SentencePiece, Marian, etc.).
	Close()
}
