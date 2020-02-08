package bandwidth

func NewLimiter(serverRate *RateConfig, connRate *RateConfig)  {

	serverLimiter := NewBandwidthLimiter(serverRate)

	// for each conn do
	RateLimitedConn(ctx, serverLimiter.Child(connRate), conn)

}
