package CReply
import (
	"testing"
	"github.com/EyciaZhou/usth-golang/M/usth"
	"strconv"
	"fmt"
	"sync"
	"math/rand"
	"runtime"
)

func TestGetScore(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	tasks := make(chan int)

	wg := sync.WaitGroup{}

	for chan_i := 0; chan_i < 100; chan_i++ {
		go func(id int) {
			for i := range tasks {
				raw, err := usth.DBScore.GetPassing(strconv.Itoa(i), "1")
				if err != nil {
					fmt.Println(id, i, "Error when getpassing" + err.Error())
					fmt.Println(id, i, (string)(raw))
					wg.Done()
					continue
				}
				name, err := usth.DBInfo.GetName(strconv.Itoa(i))
				if err != nil {
					fmt.Println(id, i, "Error when getname" + err.Error())
					wg.Done()
					continue
				}
				fmt.Println(id, i, name)
				wg.Done()
			}
		}(chan_i)
	}

	for i := 2013025001; i <= 2013025999; i++ {
		wg.Add(1)
		tasks <- i
	}
	wg.Wait()

	close(tasks)
}

func randomUsername() string {
	return strconv.Itoa(rand.Int() % 100 + 2013025000)
}

func TestReply(t *testing.T) {
	lstid := ""
	username := ""

	for i := 0; i <= 300; i++ {
		username = randomUsername()
		name, err := usth.DBInfo.GetName(username)

		for err != nil {
			username = randomUsername()
			name, err = usth.DBInfo.GetName(username)
		}

		var (
			id string
		)

		fmt.Println(i, username, name)

		if (lstid != "" && rand.Float32() < 0.3) {
			id, err = usth.DBReply.ReplyWithRef(name, username, "中 文 relpy回复 回复%%$#%@!%$^", "testclass", lstid)
		} else {
			id, err = usth.DBReply.Reply(name, username, "中 文 adsfadslkj%%$#%@!%$^", "testclass")
		}
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		lstid = id
	}
}