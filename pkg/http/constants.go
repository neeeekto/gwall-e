package http

import "net/http"

// nonRepeatableErrorStatuses содержит статус-коды, которые НЕ должны считаться повторяемыми ошибками
// и не должны приводить к открытию circuit breaker
var nonRepeatableErrorStatuses = map[int]struct{}{
	http.StatusInternalServerError: {}, // 500
	http.StatusBadGateway:          {}, // 502
	http.StatusServiceUnavailable:  {}, // 503
	http.StatusRequestTimeout:      {}, // 408
	http.StatusTooManyRequests:     {}, // 429
}