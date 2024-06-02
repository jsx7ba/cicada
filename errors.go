package cicada

import "errors"

var ErrorNotFound error = errors.New("not found")

var ErrorBadRequest error = errors.New("bad request")
