package db

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	//_ "github.com/sijms/go-ora/v2"
)

// 创建mysql连接
func connectToMySQL(dbtype, datasource string) (*sql.DB, error) {
	//if dbtype == "oracle" {
	//	//err := os.Setenv("NLS_LANG", "AMERICAN_AMERICA.ZHS16GBK")
	//	err := os.Setenv("NLS_LANG", "AMERICAN_AMERICA.AL32UTF8")
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	dbconn, err := sql.Open(dbtype, datasource)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	err = dbconn.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return dbconn, nil
}

// 连接mysql并执行sql
func ConnMysqlAndRun(dbtype string, datasource string) error {
	//创建mysql连接
	db, err := connectToMySQL(dbtype, datasource)
	if err != nil {
		return err
	}
	defer db.Close()
	reader := bufio.NewReader(os.Stdin)
	//获取用户输入的sql
	for {
		exit, err := executeUserInputSQL(db, reader)
		if err != nil {
			fmt.Printf("Error executing SQL: %v\n", err)
		} else {
			fmt.Println("SQL executed successfully.")
		}

		if exit {
			break // 用户请求退出，跳出循环
		}
	}
	return nil
}

// 执行sql
func executeUserInputSQL(db *sql.DB, reader *bufio.Reader) (bool, error) {
	fmt.Println("Enter an SQL statement (or 'exit' to exit): ")
	fmt.Print(">")
	userInput, err := reader.ReadString(';')
	// 去除输入的头部、尾部的空格及换行符
	userInput = strings.TrimSpace(userInput)
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}
	// exit退出
	//if userInput == "exit" {
	//	return true, nil
	//}
	if strings.HasPrefix(userInput, "exit") {
		return true, nil
	}
	// 提取sql语句，并以空格分割
	userInputCommand := strings.Split(userInput, " ")
	// 判断是否是source
	if len(userInputCommand) == 2 {
		userInputComm := userInputCommand[0]
		userInputArgs := userInputCommand[1]
		if userInputComm == "source" {
			// 读取source文件
			file, err := os.Open(userInputArgs)
			if err != nil {
				return false, fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()
			readfile := bufio.NewScanner(file)
			// 使用builder拼接文件中的sql语句
			sql := strings.Builder{}
			// 逐行读取文件
			for readfile.Scan() {
				// 去除sql的行首、行尾的空格及换行符
				line := strings.TrimSpace(readfile.Text())
				// 拼接sql
				sql.WriteString(line)
				// 检查sql是否为注释，如果是则清空
				check := checkSql(sql.String())
				if !check {
					sql = strings.Builder{}
				}
				// 检查sql是否以分号结尾，如果是则执行
				if strings.HasSuffix(line, ";") {
					fmt.Println(sql.String())
					err := executeSql(db, sql.String())
					sql = strings.Builder{}
					if err != nil {
						return false, err
					}
				}
			}
		} else {
			err := executeSql(db, userInput)
			if err != nil {
				return false, err
			}
		}
	} else {
		err := executeSql(db, userInput)
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

// 执行sql并输出结果
func executeSql(db *sql.DB, userInput string) error {
	rows, err := db.Query(userInput)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to retrieve columns: %w", err)
	}

	if len(cols) > 0 {
		for _, i := range cols {
			fmt.Printf(" - %s - ", i)
		}
		fmt.Println()
	}

	values := make([]interface{}, len(cols))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		//for i, col := range cols {
		//	val := values[i]
		//	switch v := val.(type) {
		//	case []byte:
		//		fmt.Printf("%s: %s\t", col, string(v))
		//	default:
		//		fmt.Printf("%s: %v\t", col, v)
		//	}
		//}
		for i, _ := range cols {
			val := values[i]
			switch v := val.(type) {
			case []byte:
				fmt.Printf(" %s\t", string(v))
			default:
				fmt.Printf(" %v\t", v)
			}
		}
		fmt.Println()
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error during rows iteration: %w", err)
	}
	return nil
}
