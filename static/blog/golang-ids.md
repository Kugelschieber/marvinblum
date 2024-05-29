**Published on 7. July 2020**

> _This post was originally published on Medium a while ago. I just added it to my blog for completeness._

![Gopher](/static/blog/hideid/gopher.png)

Most Golang web applications use persistency in some way or another. Usually, the connection between your application and the persistent layer is a technical identification value (ID), a number in most cases. IDs are useful to identify, connect and distinguish data records. Here is a typical example of a database model represented as a struct within Golang applications:

```
type Customer struct {
    Id       int64  `json:"id"`
    Email    string `json:"email"`
    Username string `json:"username"`
}
```


This struct can easily be used to retrieve and store customers in a database as well as handling customer data within your business logic. What if we add a REST endpoint to show the customers data on a website?

```
router.HandleFunc("/customer", func(w http.ResponseWriter, r* http.Request) {
    customer := findCustomer(r)
    response, _ := json.Marshal(customer)
    w.Write(response)
})
```


Calling this endpoint will return the customer object as JSON within the body:

```
{
    "id": 123,
    "email": "foo@bar.com",
    "username":"foobar"
}
```


As you can see, we received the customer object as expected. There is the email address, the username and the ID, which can be used to perform certain actions, like updating the customers username with a PUT request. We modify our endpoint to do so:

```
router.HandleFunc("/customer", func(w http.ResponseWriter, r* http.Request) {
    if r.Method == "GET" {
        customer := findCustomer(r)
        response, _ := json.Marshal(customer)
        w.Write(response)
    } else if r.Method == "PUT" {
        request := struct {
            UserId   int64  `json:"id"`
            Username string `json:"username"`
        }{}
        decoder := json.NewDecoder(r.Body)
        if err := decoder.Decode(&request); err != nil {
            w.WriteHeader(http.StatusBadRequest)
        }
        if err := updateCustomer(request.UserId, request.Username); err != nil {
            w.WriteHeader(http.StatusBadRequest)
        }
    }
})
```


Our handler accepts two methods now: GET and PUT. GET will return the customer, just like before. PUT reads the body send with the request and passes the parameters to a function updating the customer. As you can see we’ve used the ID field to identify the customer. This is a nice and simple approach to identify the customer again. So, what’s bad about all of this?

First of all: If your IDs are generated auto-incremented numbers, from a security standpoint, it’s fine to expose IDs to anyone.

On the other side: You probably don’t want to show users long boring numbers, that are hard to remember. YouTube for example uses short strings to represent a video: ?v=hY7m5jjJ9mM. This representation does not only look better in the URL, but it also hides technical IDs within their system. Another reason might be, that you don’t want to show how many records of an object exist if you use auto-incremented numbers starting at one. There are more reasons to hide technical IDs from your users, like splitting ID ranges, migrations, and so on. But I don’t want to go into too much detail here.

Take a look at this nice article by John Topley why you shouldn’t expose IDs to your users.

[

Database IDs Have No Place In URIs

https://johntopley.com/2008/08/19/database-ids-have-no-place-in-uris/



](https://johntopley.com/2008/08/19/database-ids-have-no-place-in-uris/)

At this point, it should be clear we’re looking for a simple and flexible solution for this issue. But how can we transform our IDs to a more user-friendly representation without changing too much of our existing code? The solution to this (as often in Golang): interfaces.

Instead of using int64 as our ID type, we can establish our own type and implement the interfaces needed to transform IDs into a different form. Since this article is about web applications, I assume there is (un-)marshaling to and from JSON, a database and business logic that deals with IDs. The approach I’m about to show you works for all kinds of requirements.

First of all, we declare a custom ID type:

```
type ID int64
```


As you can see, this is a simple one liner. And actually just a fancy name for an int64. In our application we want this to be returned as a hash string to the user - representing the same number - but still be a number when dealing with it internally. We have to attach a few methods to make it work.

Let’s beginn by satisfying the [Marshaler](https://golang.org/pkg/encoding/json/#Marshaler) and [Unmarshaler](https://golang.org/pkg/encoding/json/#Unmarshaler) interfaces of the standard library first:

```
// MarshalJSON implements the encoding json interface.
func (this ID) MarshalJSON() ([]byte, error) {
    if this == 0 {
        return json.Marshal(nil)
    }    result, err := hash.Encode(this)    if err != nil {
        return nil, err
    }    return json.Marshal(string(result))
}// UnmarshalJSON implements the encoding json interface.
func (this *ID) UnmarshalJSON(data []byte) error {
    // convert null to 0
    if strings.TrimSpace(string(data)) == "null" {
        *this = 0
        return nil
    }    // remove quotes
    if len(data) >= 2 {
        data = data[1 : len(data)-1]
    }    result, err := hash.Decode(data)    if err != nil {
        return err
    }    *this = ID(result)
    return nil
}// remove quotes
if len(data) >= 2 {
    data = data[1 : len(data)-1]
}result, err := hash.Decode(data)if err != nil {
    return err
}*this = ID(result)
    return nil
}
```


By adding these two methods, our ID type now translates to a hash string when it is marshalled and will be converted back to its integer representation when unmarshalled. Of course, in order for this to work, our hash function must be symmetric. You can use [HashIds](https://github.com/speps/go-hashids) for example.

Within the PUT endpoint, we can now replace the ID in the request object with our custom type:

```
request := struct {
    UserId   ID     `json:"id"`
    Username string `json:"username"`
}{}
```


Appart from that, you have to change the parameter in the updateCustomer function or cast it to an int64:

```
updateCustomer(int64(request.UserId), request.Username)
```


All that’s left to do now, is implementing the [Scanner](https://golang.org/pkg/database/sql/#Scanner) and [Valuer](https://golang.org/pkg/database/sql/driver/#Valuer) interface to persist our custom ID type within databases:

```
// Scan implements the Scanner interface.
func (this *ID) Scan(value interface{}) error {
	if value == nil {
		*this = 0
		return nil
	}

	id, ok := value.(int64)

	if !ok {
		return errors.New("unexpected type")
	}

	*this = ID(id)
	return nil
}

// Value implements the driver Valuer interface.
func (this ID) Value() (driver.Value, error) {
	return int64(this), nil
}
```


As you can see this is as simple as converting to int64, because the database driver expects all types to satisfy the [Value](https://golang.org/pkg/database/sql/driver/#Value) interface. We can now change the type of our customer ID to complete our changes:

```
type Customer struct {
    Id       ID     `json:"id"`
    Email    string `json:"email"`
    Username string `json:"username"`
}
```


And that’s it! If you want to know more about how to implement this or just use it right away, you can visit the GitHub project, which implements all of the functionality I’ve shown above. It uses HashIds, which I’ve mentioned earlier, to transform the IDs to a nice and short hash representation.
