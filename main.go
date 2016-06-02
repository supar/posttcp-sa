package main

import (
	"database/sql"
	"flag"
	"net"
)

func init() {
	log = NewLogger(1000)
	// Set console log as default
	log.SetLogger("console", `{"level":7}`)
}

func main() {
	// Close logger
	defer log.Close()

	var (
		cfg    *Config
		conn   *net.TCPConn
		db     *sql.DB
		err    error
		server *Server
		stmt   *sql.Stmt
	)

	// Read flags
	flag.Parse()

	if PrintVersion {
		showVersion()
	}

	// Read configuration
	if cfg, err = NewConfig(CONFIGFILE); err != nil {
		log.Critical(err.Error())
	} else {
		if err = cfg.Parse(); err != nil {
			log.Critical(err.Error())
		}
	}

	// Reconfigure log adapters
	for _, a := range []string{"file", "console"} {
		if err = log.SetLogAdapter(cfg, a); err != nil {
			log.Error(err)
			err = nil
		}
	}

	// Prepare statement
	if db, err = openDB(cfg.Database.DSN()); err != nil {
		log.Critical(err.Error())
	} else {
		if stmt, err = OpenStmt(db); err != nil {
			log.Critical(err.Error())
		}
	}

	// Listen on TCP port 2000 on all interfaces.
	if server, err = NewServer(cfg.Server); err != nil {
		log.Fatal(err)
	}

	defer func() {
		stmt.Close()
		db.Close()
		server.Close()
		log.Close()
	}()

	// Inform that service is started
	log.Info("Service %s started at %s", NAME, server.Addr())

	for {
		// Wait for a connection.
		if conn, err = server.AcceptTCP(); err != nil {
			log.Error(err)
		}

		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c *net.TCPConn) {
			var (
				conn    = NewConnection(c)
				fn_wrap = callback_wrap(conn, stmt, cfg)
			)

			defer func(conn *Connection) {
				log.Debug("[%s] Close from %s", conn.Id(), conn.RemoteAddr())

				conn.Close()
			}(conn)

			log.Info("[%s] Connection from %s at %s", conn.Id(), conn.RemoteAddr(), conn.LocalAddr().String())

			if err := readConn(conn, fn_wrap); err != nil {
				log.Error("[%s] %s", conn.Id(), err.Error())
			}
		}(conn)
	}
}

func callback_wrap(conn *Connection, stmt *sql.Stmt, cfg *Config) (fn func(buf []byte) (int64, error)) {
	var (
		id       = conn.Id()
		interval = cfg.GetScoreInterval()
		limit    = cfg.GetScoreLimit()
	)

	fn = func(buf []byte) (count int64, err error) {
		var (
			cmd *Command
		)

		log.Debug("[%s] Read from connection %d bytes", id, len(buf))

		if cmd, err = NewCommand(buf); err != nil {
			log.Error("[%s] %s", id, err.Error())

			return 0, err
		}

		log.Debug("[%s] Requested `%s'", id, cmd.GetStr())

		if cmd.wr {
			log.Warn("[%s] Command `put' is not supported yet", id)

			return 0, nil
		}

		log.Debug("[%s] Call prepared statement with args: client=%s, interval=%d, score_limit=%f", id, cmd.GetStr(), interval, limit)

		count, err = isSpammer(stmt, cmd.GetStr(), interval, interval, limit)

		if count == 0 {
			log.Debug("[%s] Requested `%s' was not found", id, cmd.GetStr())
		} else {
			log.Info("[%s] Requested `%s' has bothersome score %d, reject", id, cmd.GetStr(), count)
		}

		if err != nil {
			log.Error("[%s], %s", id, err.Error())
		}

		return
	}

	return
}
