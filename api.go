package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddress string
	store         Storage
}

func newAPIServer(listenAddress string, store Storage) *APIServer {
	return &APIServer{
		listenAddress: listenAddress,
		store:         store,
	}
}

func (server *APIServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/login", makeHTTPHandleFunc(server.handleLogin))
	router.HandleFunc("/account", makeHTTPHandleFunc(server.handleAccount))
	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(server.handleGetAccountByID), server.store))
	router.HandleFunc("/transfer", withJWTAuth(makeHTTPHandleFunc(server.handleTransfer), server.store))
	router.HandleFunc("/deactivate", withJWTAuth(makeHTTPHandleFunc(server.handleDeactivateAccount), server.store))
	log.Println("JSON API Server iniciado na porta: ", server.listenAddress)
	http.ListenAndServe(server.listenAddress, router)
}

// 98986
func (server *APIServer) handleLogin(writer http.ResponseWriter, request *http.Request) error {
	if request.Method != "POST" {
		return fmt.Errorf("Método não permitido %s", request.Method)
	}
	var loginRequest LoginRequest
	if err := json.NewDecoder(request.Body).Decode(&loginRequest); err != nil {
		return err
	}

	account, err := server.store.GetAccountByNumber(int(loginRequest.Number))
	if err != nil {
		return err
	}

	if !account.ValidaPassword(loginRequest.Password) {
		return fmt.Errorf("erro de autenticação")
	}

	tokenString, err := createJWT(account)
	if err != nil {
		return err
	}

	if account.Status == "Inactive" {
		if err := server.store.UpdateAccount("status", "Active", account); err != nil {
			return err
		}
	}

	response := LoginResponse{
		Token:  tokenString,
		Number: account.Number,
	}

	fmt.Printf("%+v\n", account)

	return WriteJSON(writer, http.StatusOK, response)
}

func (server *APIServer) handleAccount(writer http.ResponseWriter, request *http.Request) error {
	if request.Method == "GET" {
		return server.handleGetAccount(writer, request)
	}
	if request.Method == "POST" {
		return server.handleCreateAccount(writer, request)
	}

	return fmt.Errorf("Método não permitido %s", request.Method)
}

func (server *APIServer) handleGetAccount(writer http.ResponseWriter, request *http.Request) error {
	accounts, err := server.store.GetAccounts()
	if err != nil {
		return err
	}
	return WriteJSON(writer, http.StatusOK, accounts)
}

func (server *APIServer) handleGetAccountByID(writer http.ResponseWriter, request *http.Request) error {
	if request.Method == "GET" {
		id, err := getID(request)
		if err != nil {
			return err
		}
		account, err := server.store.GetAccountByID(id)
		if err != nil {
			return err
		}

		return WriteJSON(writer, http.StatusOK, account)
	}

	if request.Method == "DELETE" {
		return server.handleDeleteAccount(writer, request)
	}

	return fmt.Errorf("Método não permitido %s", request.Method)
}

func (server *APIServer) handleCreateAccount(writer http.ResponseWriter, request *http.Request) error {
	createAccountRequest := new(CreateAccountRequest)
	if err := json.NewDecoder(request.Body).Decode(createAccountRequest); err != nil {
		return err
	}

	account, err := NewAccount(createAccountRequest.FirstName, createAccountRequest.LastName, createAccountRequest.PasswordEncrypted)
	if err != nil {
		return err
	}
	if err := server.store.CreateAccount(account); err != nil {
		return nil
	}

	// tokenString, err := createJWT(account)
	// if err != nil {
	// 	return err
	// }

	// fmt.Println("JWT token: ", tokenString)

	return WriteJSON(writer, http.StatusOK, account)
}

func (server *APIServer) handleDeleteAccount(writer http.ResponseWriter, request *http.Request) error {
	id, err := getID(request)
	if err != nil {
		return err
	}
	if err := server.store.DeleteAccount(id); err != nil {
		return err
	}
	return WriteJSON(writer, http.StatusOK, map[string]int{"deleted": id})
}

func (server *APIServer) handleDeactivateAccount(writer http.ResponseWriter, request *http.Request) error {
	var loginRequest LoginRequest
	if err := json.NewDecoder(request.Body).Decode(&loginRequest); err != nil {
		return err
	}

	account, err := server.store.GetAccountByNumber(int(loginRequest.Number))
	if err != nil {
		return err
	}

	if err := server.store.UpdateAccount("status", "Inactive", account); err != nil {
		return err
	}

	return WriteJSON(writer, http.StatusOK, map[string]string{"message": "Conta desativada com sucesso"})
}

func (server *APIServer) handleTransfer(writer http.ResponseWriter, request *http.Request) error {
	transferRequest := new(TransferRequest)
	if err := json.NewDecoder(request.Body).Decode(transferRequest); err != nil {
		return err
	}

	defer request.Body.Close()

	return WriteJSON(writer, http.StatusOK, transferRequest)
}

func WriteJSON(writer http.ResponseWriter, status int, value any) error {
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(status)
	return json.NewEncoder(writer).Encode(value)
}

func createJWT(account *Account) (string, error) {
	claims := &jwt.MapClaims{
		"expires_at":     jwt.NewNumericDate(time.Unix(1516239022, 0)),
		"account_number": account.Number,
	}

	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func withJWTAuth(handlerFunc http.HandlerFunc, storage Storage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("Chamando JWT Auth Middleware")

		tokenString := request.Header.Get("x-jwt-token")
		token, err := validaJWT(tokenString)
		if err != nil || !token.Valid {
			permissaoNegada(writer)
			return
		}

		userID, err := getID(request)
		if err != nil {
			permissaoNegada(writer)
			return
		}

		account, err := storage.GetAccountByID(userID)
		if err != nil {
			permissaoNegada(writer)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		// panic(reflect.TypeOf(claims["accountNumber"])) // float64 ?
		if account.Number != int64(claims["account_number"].(float64)) {
			permissaoNegada(writer)
		}

		handlerFunc(writer, request)
	}
}

func permissaoNegada(writer http.ResponseWriter) {
	WriteJSON(writer, http.StatusForbidden, ApiError{Error: "Permissão negada"})
}

func validaJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Método de entrada inesperado: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if err := f(writer, request); err != nil {
			WriteJSON(writer, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func getID(request *http.Request) (int, error) {
	idString := mux.Vars(request)["id"]
	id, err := strconv.Atoi(idString)
	if err != nil {
		return id, fmt.Errorf("ID concedido é inválido %s", idString)
	}
	return id, nil
}
