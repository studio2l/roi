package roi

import (
	"fmt"
	"image"
	"image/png"
	"os"
)

// AddThumbnail은 특정 샷의 썸네일을 등록한다.
// 썸네일은 roi안에 파일로 저장된다.
func AddThumbnail(prj, shot, thumbf string) error {
	// wrap은 AddThumbnail에서 에러가 났을 때 에러 내용에 기본적인 정보를 추가한다.
	wrap := func(err error) error {
		return fmt.Errorf("AddThumbnail: %s", err)
	}

	fi, err := os.Stat(thumbf)
	if err != nil {
		return wrap(err)
	}
	maxKB := int64(200)
	if fi.Size() > (maxKB << 10) {
		return wrap(fmt.Errorf("file size is bigger than %dKB: %s", maxKB, thumbf))
	}
	from, err := os.Open(thumbf)
	if err != nil {
		return wrap(err)
	}
	defer from.Close()
	// thumbf가 지원하는 이미지 파일이 맞는지 확인한다.
	img, _, err := image.Decode(from)
	if err != nil {
		return wrap(err)
	}
	// 이미지를 png 이미지로 변경한다.
	// 파일을 부를때 일일이 파일 확장자를 검사하지 않기 위함이다.
	if err := os.MkdirAll(fmt.Sprintf("roi-userdata/thumbnail/%s", prj), 0755); err != nil {
		if !os.IsExist(err) {
			return wrap(err)
		}
	}
	to, err := os.Create(fmt.Sprintf("roi-userdata/thumbnail/%s/%s.png", prj, shot))
	if err != nil {
		return wrap(err)
	}
	defer to.Close()
	if err := png.Encode(to, img); err != nil {
		return wrap(err)
	}
	return nil
}
