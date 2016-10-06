package datastore

import (
	"log"

	"github.com/garyburd/redigo/redis"
)

type Datastore interface {
	Save(userID int, cpf, pwd string) error
	Retrieve(userID int) (cpf, pwd string, err error)
	Close() error
}

func NewRedis(url string) Datastore {
	return &redisDS{
		&redis.Pool{
			MaxIdle:   30,
			MaxActive: 30,
			Wait:      true,
			Dial: func() (redis.Conn, error) {
				log.Println("Connecting", url)
				conn, err := redis.DialURL(url)
				if err != nil {
					log.Panic("Could not connect to redis. Cause: " + err.Error())
					return nil, err
				}
				return conn, err
			},
		},
	}
}

type redisDS struct {
	pool *redis.Pool
}

func (r *redisDS) Close() error {
	return r.pool.Close()
}

func (r *redisDS) Save(userID int, cpf, pwd string) error {
	id := string(userID)
	conn := r.pool.Get()
	defer conn.Close()
	if _, err := conn.Do("SET", id+".cpf", cpf); err != nil {
		return err
	}
	_, err := conn.Do("SET", id+".pwd", pwd)
	return err
}

func (r *redisDS) Retrieve(userID int) (cpf, pwd string, err error) {
	id := string(userID)
	conn := r.pool.Get()
	defer conn.Close()
	cpf, err = redis.String(conn.Do("GET", id+".cpf"))
	if err != nil {
		return cpf, pwd, err
	}
	pwd, err = redis.String(conn.Do("GET", id+".pwd"))
	return cpf, pwd, err
}
