// appendFieldInput은 가까운 부모 field에 첫번째 인풋과 같은 형식의 새 인풋을 추가한다.
function appendFieldInput(el) {
	let field = el.closest(".field")
	// 숨겨진 첫번째 인풋은 생성될 타입과 이름을 가진 레퍼런스 인풋이다.
	let refInput = field.getElementsByTagName("input")[0]
	let input = document.createElement("input")
	input.type = refInput.type
	input.name = refInput.name
	input.value = ""
	input.style.marginTop = "0.5rem"
	field.appendChild(input)
}

// autocomplete takes input tag and possible autocompleted values and label.
function autocomplete(inp, vals, label) {
	// turn-off browser's default autocomplete behavior
	inp.setAttribute("autocomplete", "off")

	// focus is index of focused div.
	let focus = -1

	inp.addEventListener("input", function(e) {
		// someone write to this input
		closeAllLists()

		let search = this.value
		if (!search) {
			return false
		}

		focus = -1

		let a = document.createElement("div")
		a.setAttribute("id", this.id + "autocomplete-list")
		a.setAttribute("class", "autocomplete-items")
		this.parentNode.appendChild(a)

		for (let v of vals) {
			let n = search.length
			let vPre = v.substr(0, n)
			let vPost = v.substr(n)
			let l = ""
			if (label && label[v]) {
				l = label[v]
			}
			let lPre = l.substr(0, n)
			let lPost = l.substr(n)
			let vMatch = vPre.toLowerCase() == search.toLowerCase()
			let lMatch = lPre.toLowerCase() == search.toLowerCase()
			if (vMatch || lMatch) {
				let b = document.createElement("div")
				// make the matching letters bold
				if (vMatch) {
					b.innerHTML = "<strong>" + vPre + "</strong>" + vPost
					if (l) {
						 b.innerHTML += " (" + l + ")"
					}
				}
				if (lMatch) {
					b.innerHTML = v + " (<strong>" + lPre + "</strong>" + lPost + ")"
				}
				b.innerHTML += "<input type='hidden' value='" + v + "'>"
				b.addEventListener("click", function(e) {
					inp.value = this.getElementsByTagName("input")[0].value
					closeAllLists()
				})
				a.appendChild(b)
			}
		}
	})

	inp.addEventListener("keydown", function(e) {
		if (e.key.length == 1) {
			// should only handle non-character keydown event.
			// note, still there is a char key it's length is bigger than 1
			// like unicode emoji symbol. But will not handle that.
			return false
		}
		let a = document.getElementById(this.id + "autocomplete-list")
		if (!a) {
			return false
		}
		let bs = a.getElementsByTagName("div")
		if (e.key == "ArrowDown") {
			focus++
			setActive(bs)
		} else if (e.key == "ArrowUp") {
			focus--
			setActive(bs)
		} else if (e.key == "Enter") {
			e.preventDefault()
			if (bs.length == 0) {
				return
			}
			if (focus == -1) {
				focus = 0
			}
			bs[focus].click()
		}
	})

	function setActive(bs) {
		if (!bs) {
			return false
		}
		for (let b of bs) {
			b.classList.remove("autocomplete-active")
		}
		if (focus >= bs.length) {
			focus = 0
		}
		if (focus < 0) {
			focus = bs.length - 1
		}
		bs[focus].classList.add("autocomplete-active")
	}

	function closeAllLists() {
		var as = document.getElementsByClassName("autocomplete-items")
		for (let a of as) {
			a.parentNode.removeChild(a)
		}
	}

	document.addEventListener("click", function (e) {
		// user clicks else where
		closeAllLists()
	})
}

function rfc3339(d) {
    function pad(n) {
        return n < 10 ? "0" + n : n;
    }
    function timezoneOffset(offset) {
        var sign;
        if (offset === 0) {
            return "Z";
        }
        sign = (offset > 0) ? "-" : "+";
        offset = Math.abs(offset);
        return sign + pad(Math.floor(offset / 60)) + ":" + pad(offset % 60);
    }
    return d.getFullYear() + "-" +
        pad(d.getMonth() + 1) + "-" +
        pad(d.getDate()) + "T" +
        pad(23) + ":" +
        pad(59) + ":" +
        pad(59) +
        timezoneOffset(d.getTimezoneOffset());
}
