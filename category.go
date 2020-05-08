package roi

import (
	"fmt"
)

// verifyCategoryName은 받아들인 카테고리 이름이 유효하지 않다면 에러를 반환한다.
func verifyCategoryName(ctg string) error {
	switch ctg {
	case "shot", "asset":
		return nil
	}
	return fmt.Errorf("invalid category: %s", ctg)
}

func verifyCategoryPrimaryKeys(show, ctg string) error {
	err := verifyShowName(show)
	if err != nil {
		return err
	}
	err = verifyCategoryName(ctg)
	if err != nil {
		return err
	}
	return nil
}
