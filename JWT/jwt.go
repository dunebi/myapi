package JWT

import (
	"errors"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

/* JWT는 Header, Payload, Signature로 이루어짐 */
/* Payload는 "전송하는 내용" 정도의 뜻을 담고 있음. 로그인 정보라고 볼 수 있음 */
type LoginPayload struct {
	Account_Id  string `json:"account_id"`
	Account_Pwd string `json:"account_pwd"`
}

type LoginResponse struct {
	Token string `json:"token"` // 로그인 시 응답으로 받는 Token, JWT Token이 될 것으로 보임
}

type JwtClaim struct {
	Account_Id string
	jwt.StandardClaims
}

/* 로그인 후 사용할 JWT 토큰을 생성함 */
func GenerateToken(account_id string) (signedToken string, err error) {
	claims := &JwtClaim{ // Account ID와 만료에 대한 정보를 담고 있음
		Account_Id: account_id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(15 * time.Minute).Unix(), // 15분 뒤 만료
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)   // token 생성
	signedToken, err = token.SignedString([]byte("SecreteCode")) // Secretecode는 토큰 자체 인증키(자체 비밀번호 느낌)

	if err != nil {
		return
	}
	return
}

/* JWT 토큰 검증 */
func ValidateToken(signedToken string) (claims *JwtClaim, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JwtClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte("SecreteCode"), nil
		},
	)

	if err != nil {
		return
	}

	claims, ok := token.Claims.(*JwtClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("jwt is expired")
		return
	}
	return
}