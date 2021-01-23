package app_lib

import (
	"context"
	"github.com/buildboxapp/app/pkg/model"
	"sort"
	"strconv"
	"strings"
	"net/http"
	"encoding/json"
	"crypto/sha1"
	"encoding/hex"
	"os"
	"net/url"
	"io/ioutil"
	"fmt"
	"bytes"
	"time"
	"html/template"
	"path"
	"sync"
	"github.com/labstack/gommon/log"
	"errors"
)

// инициируем встроенные функции для объекта приложения
// если делать просто в init, то в gui добавленные фукнции не будут видны




// удаляем элемент из слайса
func (p *model.ResponseData) RemoveData(i int) bool {

	if (i < len(p.Data)){
		p.Data = append(p.Data[:i], p.Data[i+1:]...)
	} else {
		//log.Warning("Error! Position invalid (", i, ")")
		return false
	}

	return true
}


