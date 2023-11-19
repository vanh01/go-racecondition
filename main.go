package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type User struct {
	id   int
	name string
}

type Product struct {
	id       int
	name     string
	quantity int
}

type Ordering struct {
	id        int
	userId    int
	productId int
}

var Db *sql.DB

func InitDb() *sql.DB {
	connStr := "user=ordering-system dbname=postgres password=123456 port=5432 host=localhost search_path=public sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func CheckProduct(id, quantity int) bool {
	rows, err := Db.Query(`SELECT * FROM product WHERE id = $1 AND quantity >= $2`, id, quantity)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer rows.Close()
	return rows.Next()
}

func DecreaseQuantityByProductId(product_id, quantity int) error {
	// This statement will raise an error if current quantity < quantity base on constrain on product table
	_, err := Db.Exec(`UPDATE product SET quantity=quantity-$1 WHERE id= $2;`, quantity, product_id)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func IncreaseQuantityByProductId(product_id, quantity int) error {
	// This statement will raise an error if current quantity < quantity base on constrain on product table
	_, err := Db.Exec(`UPDATE product SET quantity=quantity+$1 WHERE id= $2;`, quantity, product_id)
	if err != nil {
		return err
	}
	return nil
}

func CreateNewOrdering(user_id, product_id, quantity int) error {
	_, err := Db.Exec(`INSERT INTO ordering (user_id, product_id, quantity) VALUES($1, $2, $3);`, user_id, product_id, quantity)
	if err != nil {
		return err
	}
	return nil
}

func OrderingProcess(user_id, product_id, quantity int, wg *sync.WaitGroup, i chan int) {
	defer wg.Done()
	i <- 1
	fmt.Println(user_id, product_id, quantity)
	// get quantity from product table
	fmt.Println("Get quantity stage")
	ok := CheckProduct(product_id, quantity)
	if !ok {
		<-i
		return
	}
	fmt.Println("The product isn't yet sold out")
	// Wait for all users to finish reading product data

	fmt.Println("Update quantity stage")
	if DecreaseQuantityByProductId(product_id, quantity) != nil {
		fmt.Println("Failed to update quantity")
		<-i
		return
	}

	// create ordering if updated quantity successfully
	fmt.Println("Create ordering stage")
	if CreateNewOrdering(user_id, product_id, quantity) != nil {
		fmt.Println("Failed to create ordering, need to roll back quantity of product")
		if DecreaseQuantityByProductId(product_id, quantity) != nil {
			fmt.Println("Failed to roll back quantity")
			<-i
			return
		}
	}
	fmt.Println("Created ordering successfully!", <-i)
}

func PrepareData(product_id, number_of_user int) {
	fmt.Println("Check product exist")
	rows, err := Db.Query(`SELECT * FROM product WHERE id = $1;`, product_id)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if rows.Next() {
		// update quantity
		fmt.Println("Update product quantity")
		_, err = Db.Exec(`UPDATE product SET quantity=$1 WHERE id=$2;`, number_of_user-1, product_id)
		if err != nil {
			fmt.Println("Failed to update quantity")
			return
		}
	} else {
		// create product
		fmt.Println("Insert product")
		_, err = Db.Exec(`INSERT INTO product (id, "name", quantity) VALUES($1, 'a', $2);`, product_id, number_of_user-1)
		if err != nil {
			fmt.Println("Failed to create quantity")
			return
		}
	}
	rows.Close()

	fmt.Println("Check and create user")
	for i := 1; i <= number_of_user; i++ {
		rows, err := Db.Query(`SELECT * FROM "user" WHERE id = $1;`, i)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if !rows.Next() {
			// create user
			_, err = Db.Exec(`INSERT INTO "user" (id, "name") VALUES($1, $2);`, i, fmt.Sprintf("user %d", i))
			if err != nil {
				fmt.Println("Failed to create user")
				return
			}
		}
		rows.Close()
	}
	fmt.Println("Success preparing data")
}

func main() {
	Db = InitDb()
	fmt.Println("Successfully connected to the PostgreSQL database")
	Db.SetConnMaxIdleTime(30 * time.Minute)
	defer Db.Close()
	productId := 1
	quantityPerOrder := 1
	numberOfUser := 100000
	PrepareData(productId, numberOfUser)
	maxReq := make(chan int, 1000)

	var wg sync.WaitGroup
	wg.Add(numberOfUser)
	for i := 1; i <= numberOfUser; i++ {
		go OrderingProcess(i, productId, quantityPerOrder, &wg, maxReq)
	}
	wg.Wait()
}
