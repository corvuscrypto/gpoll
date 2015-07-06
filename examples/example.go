package main

import  (
        "github.com/corvuscrypto/gpoll"
        "net/http"
        "net/url"
        "encoding/json"
        "fmt"
        )

func sendMess(w http.ResponseWriter, r *http.Request){
  type data struct {
    Msg string `json:"msg"`
    User string `json:"user"`
  }
  values,_ := url.ParseQuery(r.URL.RawQuery)

  nD := &data{Msg: values.Get("msg"), User: values.Get("user")}
  msg,_ := json.Marshal(*nD)

  fmt.Println(string(msg))
  gpoll.Send(string(msg))

  w.WriteHeader(200)

}

func servePage(w http.ResponseWriter, r *http.Request){

  http.ServeFile(w, r, "C:/Users/Cliff/Documents/Go Projects/gpoll/index.html")

}

func main(){

  gpoll.ListenAndBroadcast("/goPoll")
  http.HandleFunc("/",servePage)
  http.HandleFunc("/send",sendMess)
  http.ListenAndServe(":25777",nil)


}
