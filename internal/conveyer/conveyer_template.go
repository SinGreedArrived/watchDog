package conveyer

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (self *Conveyer) AddStep(cmd []string) error {
	switch cmd[0] {
	case "download":
		self.steps = append(self.steps, func(conveyer *Conveyer) error {
			elem, err := url.Parse(string(conveyer.GetData()))
			if err != nil {
				return err
			}
			conveyer.httpRequest.URL = elem
			resp, err := conveyer.httpClient.Do(conveyer.httpRequest)
			if err != nil {
				return err
			}
			if resp.StatusCode != 200 {
				return errors.New(fmt.Sprintf("Status code: %d", resp.StatusCode))
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			conveyer.WriteData(body)
			return nil
		})
	case "remove":
		self.steps = append(self.steps, func(conveyer *Conveyer) error {
			reg := cmd[1]
			regxp, err := regexp.Compile(reg)
			if err != nil {
				return err
			}
			data := regxp.ReplaceAll(conveyer.data[conveyer.pointer], []byte(""))
			conveyer.WriteData(data)
			return nil
		})
	case "replace":
		self.steps = append(self.steps, func(conveyer *Conveyer) error {
			reg := cmd[1]
			repl := cmd[2]
			regxp, err := regexp.Compile(reg)
			if err != nil {
				return err
			}
			data := regxp.ReplaceAll(conveyer.data[conveyer.pointer], []byte(repl))
			conveyer.WriteData(data)
			return nil
		})
	case "css":
		self.steps = append(self.steps, func(conveyer *Conveyer) error {
			selector := cmd[1]
			number := 0
			if len(cmd) > 2 {
				number, _ = strconv.Atoi(cmd[2])
			}
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(conveyer.data[conveyer.pointer])))
			if err != nil {
				return err
			}
			slc := doc.Find(selector).Eq(number)
			body, err := goquery.OuterHtml(slc)
			if err != nil {
				return err
			}
			conveyer.WriteData([]byte(body))
			return nil
		})
	case "undo":
		self.steps = append(self.steps, func(conveyer *Conveyer) error {
			i, _ := strconv.Atoi(cmd[1])
			for ; i != 0; i-- {
				conveyer.pointer--
				if conveyer.pointer < 0 {
					conveyer.pointer = 0
				}
			}
			return nil
		})
	case "redo":
		self.steps = append(self.steps, func(conveyer *Conveyer) error {
			i, _ := strconv.Atoi(cmd[1])
			for ; i != 0; i-- {
				conveyer.pointer++
				if conveyer.pointer == len(conveyer.data) {
					conveyer.pointer--
				}
			}
			return nil
		})
	case "md5":
		self.steps = append(self.steps, func(conveyer *Conveyer) error {
			hash := md5.Sum(conveyer.data[conveyer.pointer])
			conveyer.WriteData([]byte(hash[:]))
			return nil
		})
	case "show":
		self.steps = append(self.steps, func(conveyer *Conveyer) error {
			fmt.Println(string(conveyer.data[conveyer.pointer]))
			return nil
		})
	default:
		return errors.New(fmt.Sprintf("Command %s unknown", cmd[0]))
	}
	return nil
}
