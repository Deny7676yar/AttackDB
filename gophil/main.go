package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if err := populateDB(); err != nil {
		log.Fatal(err)
	}
}

func populateDB() error {
	c := ReadConfig()
	if err := populateEmployees(c); err != nil {
		return fmt.Errorf("failed to populate the employees table: %w", err)
	}
	return nil
}

func populateEmployees(c *Config) error {
	emps, err := NewEmps()
	if err != nil {
		return fmt.Errorf("failed to create an emps struct: %w", err)
	}
	pool, err := createPGXPool(&c.dbConfig, int32(c.PoolMaxConns), int32(c.PoolMinConns))
	if err != nil {
		return fmt.Errorf("pool creation failed: %w", err)
	}
	defer pool.Close()
	ctx := context.TODO()
	wp := NewWorkerPool(ctx, c.GoroutinesCount)
	const chunkSize = 100
	const empsLimit = 100000
	workersSync := &sync.WaitGroup{}
	processedNum := 0
	processedMux := &sync.Mutex{}
	for i := 0; i < empsLimit/chunkSize; i++ {
		res := wp.Run(func(ctx context.Context) (err error) {
			log.Println("running new job")
			defer log.Println("job finished")
			tx, err := pool.Begin(ctx)
			if err != nil {
				return fmt.Errorf("failed to start the DB transaction: %w", err)
			}
			defer func() {
				if err != nil {
					if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
						err = fmt.Errorf("rollback failed (%v), original err: %w", rollbackErr, err)
					}
				}
				if commitErr := tx.Commit(ctx); commitErr != nil {
					err = fmt.Errorf("commit failed (%v), original err: %w", commitErr, err)
				}
			}()
			b := &pgx.Batch{}
			for i := 0; i < chunkSize; i++ {
				stmt := `INSERT INTO employees(first_name, last_name, phone, email, salary, manager_id, department, position)
							VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
				b.Queue(stmt, emps.GetEmployee()...)
			}
			if err := tx.SendBatch(ctx, b).Close(); err != nil {
				return fmt.Errorf("failed to send the batch: %w", err)
			}
			return nil
		})
		workersSync.Add(1)
		go func() {
			defer workersSync.Done()
			e := <-res
			if e != nil {
				log.Println("an error occurred: %w", e)
			}
			processedMux.Lock()
			defer processedMux.Unlock()
			processedNum += chunkSize
			if processedNum%1000 == 0 {
				log.Println("processed: ", processedNum)
			}
		}()
	}
	workersSync.Wait()
	return nil
}

type Emps struct {
	Names       []string `json:"names"`
	LastNames   []string `json:"surnames"`
	PhoneCodes  []string `json:"phoneCodes"`
	Positions   []int    `json:"positions"`
	Departments []int    `json:"departments"`
	Managers    []int    `json:"managers"`

	mux *sync.RWMutex `json:"-"`
}

func NewEmps() (*Emps, error) {
	empsContents, err := os.ReadFile("emps.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open the data file: %w", err)
	}
	var emps Emps
	if err := json.Unmarshal(empsContents, &emps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the emps data: %w", err)
	}
	emps.mux = &sync.RWMutex{}
	return &emps, nil
}

func (e *Emps) name() string {
	return getRandStrEl(e.Names)
}

func (e *Emps) lastName() string {
	return getRandStrEl(e.LastNames)
}

func (e *Emps) phone() string {
	phoneCode := getRandStrEl(e.PhoneCodes)
	const minDigit, maxDigit = 0, 9
	const phoneNumLen = 7
	phoneNum := make([]byte, 0, phoneNumLen)
	for i := 0; i < phoneNumLen; i++ {
		digit := getRandIntInInterval(minDigit, maxDigit)
		phoneNum = append(phoneNum, byte('0'+digit))
	}
	return fmt.Sprintf("+7(%s)%s", phoneCode, string(phoneNum))
}

func (e *Emps) salary() int {
	const minSalary, maxSalary = 1000, 100000
	return getRandIntInInterval(minSalary, maxSalary)
}

func (e *Emps) position() int {
	return getRandIntEl(e.Positions)
}

func (e *Emps) manager() int {
	return getRandIntEl(e.Managers)
}

func (e *Emps) department() int {
	return getRandIntEl(e.Departments)
}

func (e *Emps) email(firstName string, lastName string) string {
	return fmt.Sprintf("%s%s@gopher_corp.com", string(firstName[0]), lastName)
}

func (e *Emps) GetEmployee() []interface{} {
	e.mux.RLock()
	defer e.mux.RUnlock()

	name := e.name()
	lastName := e.lastName()
	return []interface{}{
		name,
		lastName,
		e.phone(),
		e.email(name, lastName),
		e.salary(),
		e.manager(),
		e.department(),
		e.position(),
	}
}

func getRandIntInInterval(min int, max int) int {
	return rand.Intn(max-min) + min
}

func getRandStrEl(sl []string) string {
	return sl[rand.Intn(len(sl))]
}

func getRandIntEl(sl []int) int {
	return sl[rand.Intn(len(sl))]
}

func createPGXPool(dbConf *dbConfig, maxConns int32, minConns int32) (*pgxpool.Pool, error) {
	cfg, err := getPoolConfig(dbConf, maxConns, minConns)
	if err != nil {
		return nil, fmt.Errorf("failed to get the pool config: %w", err)
	}
	pool, err := pgxpool.ConnectConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}
	return pool, nil
}

func getPoolConfig(dbConf *dbConfig, maxConns int32, minConns int32) (*pgxpool.Config, error) {
	connStr := ComposeConnectionString(dbConf)
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create a pool config from connection string %s: %w", connStr, err,
		)
	}
	cfg.MaxConns = maxConns
	cfg.MinConns = minConns

	// HealthCheckPeriod - частота проверки работоспособности
	// соединения с Postgres
	cfg.HealthCheckPeriod = 1 * time.Minute

	// MaxConnLifetime - сколько времени будет жить соединение.
	// Так как большого смысла удалять живые соединения нет,
	// можно устанавливать большие значения
	cfg.MaxConnLifetime = 24 * time.Hour

	// MaxConnIdleTime - время жизни неиспользуемого соединения,
	// если запросов не поступало, то соединение закроется.
	cfg.MaxConnIdleTime = 30 * time.Minute

	// ConnectTimeout устанавливает ограничение по времени
	// на весь процесс установки соединения и аутентификации.
	cfg.ConnConfig.ConnectTimeout = 1 * time.Second

	// Лимиты в net.Dialer позволяют достичь предсказуемого
	// поведения в случае обрыва сети.
	cfg.ConnConfig.DialFunc = (&net.Dialer{
		KeepAlive: cfg.HealthCheckPeriod,
		// Timeout на установку соединения гарантирует,
		// что не будет зависаний при попытке установить соединение.
		Timeout: cfg.ConnConfig.ConnectTimeout,
	}).DialContext
	return cfg, nil
}

func ComposeConnectionString(c *dbConfig) string {
	userspec := fmt.Sprintf("%s:%s", c.DBUser, c.DBPassword)
	hostspec := fmt.Sprintf("%s:%s", c.DBHost, c.DBPort)
	return fmt.Sprintf("postgresql://%s@%s/%s", userspec, hostspec, c.DBName)
}
