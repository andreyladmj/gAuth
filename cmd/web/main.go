package main

import (
	"andreyladmj/gAuth/pkg/grpcapi"
	"andreyladmj/gAuth/pkg/grpcapi/userpb"
	"database/sql"
	"flag"
	"github.com/golangcollege/sessions"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"andreyladmj/gAuth/pkg/models/mysql"

	_ "github.com/go-sql-driver/mysql"

	"golang.org/x/oauth2"
	googleOAuth2 "golang.org/x/oauth2/google"

	"github.com/golang-migrate/migrate/v4"
	mysql_migrate "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)



type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	session  *sessions.Session
	users *mysql.UserModel
	oauth2Config *oauth2.Config
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.LUTC|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.LUTC|log.Ltime|log.Lshortfile)

	db, err := openDB(os.Getenv("DB_CONN"))
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	usermodel := &mysql.UserModel{DB: db}
	grpcServer := &grpcapi.GRPCServer{
		UserModel:usermodel,
		ErrorLog:errorLog,
		InfoLog:infoLog,
	}

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		users:    usermodel,
		session:  makeSession(),
		oauth2Config: &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("REDIRECT_URI"),
			Endpoint:     googleOAuth2.Endpoint,
			Scopes:       []string{"profile", "email"},
		},
	}

	//tslconfig := &tls.Config{
	//	PreferServerCipherSuites: true,
	//	CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256}, // these two have assembly implementations
	//}

	server := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		//TLSConfig:    tslconfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	driver, err := mysql_migrate.WithInstance(db, &mysql_migrate.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file://pkg/models/mysql/migrations",
		"mysql", driver)

	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil {
		errorLog.Printf("Migration %s", err)
	}

	infoLog.Printf("Starting server on %s", *addr)
	go startGRPCServer(grpcServer)

	//err = server.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	err = server.ListenAndServe()
	errorLog.Fatal(err)
}


func startGRPCServer(grpcServer *grpcapi.GRPCServer) {
	grpcServer.InfoLog.Printf("GRPC start serving")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	userpb.RegisterAuthServiceServer(s, grpcServer)

	if err := s.Serve(lis); err != nil {
		grpcServer.ErrorLog.Printf("GRPC failed to serve: %v", err)
	}

	grpcServer.InfoLog.Printf("GRPC stop serving")
}

func makeSession() *sessions.Session {
	secret := StringWithCharset(32, "")
	session := sessions.New([]byte(secret))
	session.Lifetime = 10 * time.Minute
	//session.Secure = true
	session.SameSite = http.SameSiteStrictMode
	return session
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

