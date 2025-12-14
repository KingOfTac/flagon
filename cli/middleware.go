package cli

type Middleware func(next Handler) Handler

func (c *CLI) Use(m Middleware) {
	if m != nil {
		c.Middleware = append(c.Middleware, m)
	}
}

func applyMiddleware(h Handler, m []Middleware) Handler {
	if h == nil {
		return nil
	}

	for i := len(m) - 1; i >= 0; i-- {
		if m[i] == nil {
			continue
		}
		h = m[i](h)
	}
	return h
}
