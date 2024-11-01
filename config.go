package main

type Config struct {
	Port int `envconfig:"PORT" default:"5000"`
}
