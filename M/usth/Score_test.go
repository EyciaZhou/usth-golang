package usth
import "testing"

func TestPostRequest(t *testing.T) {
	f := NewFetcher()
	err := f.Login("2013025014", "1")
	//err := f.All()
	if err != nil {
		panic(err)
	}

	//info, err := f.SchoolRollInfo()
	//if err != nil {
	//	panic(err)
	//}
}