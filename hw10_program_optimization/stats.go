package hw10programoptimization

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/mailru/easyjson"
)

//easyjson:json
type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

var (
	ErrScannerError   = errors.New("got a scanner error")
	ErrNotValidJSON   = errors.New("got not valid json string")
	ErrNotValidDomain = errors.New("can't create regexp with target domain name")
)

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	u, err := getUsers(r)
	if err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}
	return countDomains(u, domain)
}

type users [100_000]User

func getUsers(r io.Reader) (users, error) {
	var result users

	scanner := bufio.NewScanner(r)

	i := 0
	for scanner.Scan() {
		var user User
		if err := easyjson.Unmarshal(scanner.Bytes(), &user); err != nil {
			return result, ErrNotValidJSON
		}
		result[i] = user
		i++
	}

	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf("%w %s", ErrScannerError, err)
	}

	return result, nil
}

func countDomains(u users, domain string) (DomainStat, error) {
	result := make(DomainStat)

	re, err := regexp.Compile(fmt.Sprintf("\\.%s", domain))
	if err != nil {
		return nil, ErrNotValidDomain
	}

	for _, user := range u {
		matched := re.Match([]byte(user.Email))

		if matched {
			num := result[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])]
			num++
			result[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])] = num
		}
	}
	return result, nil
}
