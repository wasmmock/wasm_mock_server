package capabilities

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tebeka/selenium"
)

func SeleniumStart(wdMap *map[string]selenium.WebDriver, wdServiceMap *map[string]*selenium.Service, uid string) ([]byte, error) {
	const (
		// These paths will be different on your system.
		seleniumPath     = "../selenium_vendor/selenium/vendor/selenium-server.jar"
		geckoDriverPath  = "../selenium_vendor/selenium/vendor/geckodriver"
		chromeDriverPath = "../selenium_vendor/selenium/vendor/chromedriver"
		port             = 7771
	)
	opts := []selenium.ServiceOption{
		//	selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		selenium.ChromeDriver(chromeDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		selenium.Output(os.Stderr),              // Output debug information to STDERR.
	}
	selenium.SetDebug(true)
	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		return []byte{}, err
	}
	(*wdServiceMap)[uid] = service
	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		return []byte{}, err
	}
	(*wdMap)[uid] = wd
	return []byte{}, nil
}
func SeleniumGet(wdMap *map[string]selenium.WebDriver, uid string, addr []byte) ([]byte, error) {
	addrStr := string(addr)
	if val, ok := (*wdMap)[uid]; ok {
		err := val.Get(addrStr)
		return []byte{}, err
	}
	return []byte{}, fmt.Errorf("Can't find the WebDriver for uid: ", uid)
}
func SeleniumFindElement(wdMap *map[string]selenium.WebDriver, wdElementMap *map[string]selenium.WebElement, uid string, selector []byte) ([]byte, error) {
	selectorStr := string(selector)
	if val, ok := (*wdMap)[uid]; ok {
		elem, err := val.FindElement(selenium.ByCSSSelector, selectorStr)
		if err != nil {
			return []byte{}, err
		}
		(*wdElementMap)[uid] = elem
		return []byte{}, nil
	}
	return []byte{}, fmt.Errorf("Can't find the WebDriver for uid: %v", uid)
}
func SeleniumFindElements(wdMap *map[string]selenium.WebDriver, wdElementMap *map[string]selenium.WebElement, uid string, index int, selector []byte) ([]byte, error) {
	selectorStr := string(selector)
	if val, ok := (*wdMap)[uid]; ok {
		elems, err := val.FindElements(selenium.ByCSSSelector, selectorStr)
		if err != nil {
			return []byte{}, err
		}
		if len(elems) > index {
			(*wdElementMap)[uid] = elems[index]
		}
		return []byte{}, fmt.Errorf("Len of elements not long enough, Cannot index: %v", index)
	}
	return []byte{}, fmt.Errorf("Can't find the WebDriver for uid: %v", uid)
}
func SeleniumClick(wdElementMap *map[string]selenium.WebElement, uid string) ([]byte, error) {
	if val, ok := (*wdElementMap)[uid]; ok {
		if err := val.Click(); err != nil {
			return []byte{}, fmt.Errorf("Click err for uid: %v, %v", uid, err)
		}
		return []byte{}, nil
	}
	return []byte{}, fmt.Errorf("Can't find the WebElement for uid: %v", uid)
}
func SeleniumSendKeys(wdElementMap *map[string]selenium.WebElement, uid string, text []byte) ([]byte, error) {
	if val, ok := (*wdElementMap)[uid]; ok {
		if err := val.SendKeys(string(text)); err != nil {
			return []byte{}, fmt.Errorf("SendKeys err for uid: %v, %v", uid, err)
		}
		return []byte{}, nil
	}
	return []byte{}, fmt.Errorf("Can't find the WebElement for uid: %v", uid)
}
func SeleniumGetCookies(wdMap *map[string]selenium.WebDriver, uid string) ([]byte, error) {
	if val, ok := (*wdMap)[uid]; ok {
		if cArr, err := val.GetCookies(); err == nil {
			return json.Marshal(cArr)
		}
		return []byte{}, fmt.Errorf("Can't find the Cookies for uid: %v", uid)
	}
	return []byte{}, fmt.Errorf("Can't find the WebDriver for uid: %v", uid)
}
func SeleniumClose(wdMap *map[string]selenium.WebDriver, wdServiceMap *map[string]*selenium.Service, wdElementMap *map[string]selenium.WebElement, uid string) ([]byte, error) {
	if val, ok := (*wdMap)[uid]; ok {
		val.Quit()
		delete(*wdMap, uid)
	}
	if val, ok := (*wdServiceMap)[uid]; ok {
		val.Stop()
		delete(*wdServiceMap, uid)
	}
	if _, ok := (*wdElementMap)[uid]; ok {
		delete(*wdElementMap, uid)
	}
	return []byte{}, nil
}
