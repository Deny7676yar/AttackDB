package main

import "flag"

type Config struct {
	dbConfig
	poolConfig
}

type dbConfig struct {
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
}

type poolConfig struct {
	PoolMaxConns    int
	PoolMinConns    int
	GoroutinesCount int
}

func ReadConfig() *Config {
	c := &Config{}
	flag.StringVar(&c.DBHost, "host", "127.0.0.1", "DB host")
	flag.StringVar(&c.DBPort, "port", "5432", "DB port")
	flag.StringVar(&c.DBName, "db", "gopher_corp", "DB name")
	flag.StringVar(&c.DBUser, "user", "gopher", "DB user")
	flag.StringVar(&c.DBPassword, "password", "P@ssw0rd", "DB password")
	flag.IntVar(&c.PoolMaxConns, "max-conns", 8, "connection pool MaxConnections param")
	flag.IntVar(&c.PoolMaxConns, "min-conns", 8, "connection pool MinConnections param")
	flag.IntVar(&c.GoroutinesCount, "goroutines", 50, "number of goroutines to run")
	flag.Parse()
	return c
}
