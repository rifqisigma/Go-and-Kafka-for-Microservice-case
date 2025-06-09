package utils

import "errors"

var (
	ErrInternal         = errors.New("internal error")
	ErrInvalidEmail     = errors.New("email tidak sesuai")
	ErrNotAdmin         = errors.New("kau bukan admin")
	ErrNoStore          = errors.New("tidak ada store")
	ErrNoTopic          = errors.New("bukan ada topic ini")
	ErrFailedKafkaWrite = errors.New("gagal  mengirim message ")
	ErrUnavaible        = errors.New("tidak ada hasil")
	ErrStocknotEnough   = errors.New("stok product tidak cukup")
	ErrProductDeleted   = errors.New("product dihapus")
)
