package constrictor

import (
	"fmt"
	"strings"
)

// HTML renders a whole game as one self-contained page: a grid of colored
// squares with a slider to scrub through turns, arrow keys to step.
func HTML(frames []State, ego int) string {
	if len(frames) == 0 {
		return ""
	}
	quoted := make([]string, len(frames))
	for i, f := range frames {
		quoted[i] = `"` + strings.Join(f.rows(), "") + `"`
	}
	return fmt.Sprintf(page, frames[0].W, frames[0].H, ego, strings.Join(quoted, ","))
}

const page = `<!doctype html>
<meta charset="utf-8">
<title>constrictor</title>
<style>
body{background:#14161a;color:#c8ccd4;font:14px ui-monospace,SFMono-Regular,Menlo,monospace;
     display:flex;flex-direction:column;align-items:center;gap:14px;padding:28px}
#grid{display:grid;gap:2px}
.c{width:28px;height:28px;border-radius:3px;background:#22252b}
.ego{background:#4ea8de}
.foe0{background:#f6bd60}.foe1{background:#f4a261}.foe2{background:#e76f51}.foe3{background:#c1443b}
.head{outline:3px solid #f5f7fa;outline-offset:-3px}
#bar{display:flex;align-items:center;gap:10px}
button{background:#22252b;color:inherit;border:1px solid #343842;border-radius:4px;
       padding:4px 12px;font:inherit;cursor:pointer}
input[type=range]{width:340px}
</style>
<div id="info"></div>
<div id="grid"></div>
<div id="bar">
  <button id="prev">&larr;</button>
  <input type="range" id="slider" min="0" value="0">
  <button id="next">&rarr;</button>
</div>
<script>
const W = %d, H = %d, EGO = %d, F = [%s];
const seatClass = seat => seat === EGO ? "ego" : "foe" + seat;
const grid = document.getElementById("grid");
grid.style.gridTemplateColumns = "repeat(" + W + ",28px)";

const cells = [];
for (let i = 0; i < W * H; i++) {
  const d = document.createElement("div");
  d.className = "c";
  grid.appendChild(d);
  cells.push(d);
}

const slider = document.getElementById("slider");
slider.max = F.length - 1;

function draw() {
  const f = F[slider.value];
  for (let i = 0; i < cells.length; i++) {
    const ch = f[i];
    if (ch === ".") {
      cells[i].className = "c";
    } else if (ch >= "A" && ch <= "Z") {
      cells[i].className = "c " + seatClass(ch.charCodeAt(0) - 65) + " head";
    } else {
      cells[i].className = "c " + seatClass(ch.charCodeAt(0) - 48);
    }
  }
  info.textContent = "turn " + slider.value + " of " + (F.length - 1);
}

const step = n => { slider.value = Math.min(F.length - 1, Math.max(0, +slider.value + n)); draw(); };
slider.oninput = draw;
prev.onclick = () => step(-1);
next.onclick = () => step(1);
onkeydown = e => {
  if (e.key === "ArrowLeft") step(-1);
  if (e.key === "ArrowRight") step(1);
};
draw();
</script>
`
