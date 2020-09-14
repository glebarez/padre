package encoder

import "encoding/base64"

func NewB64encoder(replacements string) Encoder {
	return newEncoderWithReplacer(base64.StdEncoding, replacements)
}

func NewLHEXencoder(replacements string) Encoder {
	return newEncoderWithReplacer(&lhexEncoder{}, replacements)
}

func NewASCIIencoder() Encoder {
	return &asciiEncoder{}
}
