Go library for ingesting log data into CrowdStrike Falcon Logscale aka Humio using zerolog

Basis usage:

	logscale := NewLogscaleLogger("url", "token", "test")
	log.Info().Msg("test humio 1")
	log.Info().Msg("test humio 2")
	log.Info().Msg("test humio 3")
	log.Error().Msg("test error")
	logscale.WaitTillAllMessagesSend()


This impl does not support the following usage properly:

    log.Err("error msg").Msg("test humio")