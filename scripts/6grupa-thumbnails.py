"""Feature thumbnails for the 6G-RUPA cross-posts. Flat diagrams, 6grupa.com palette.
Blowfish list cards crop to ~5:3, so everything stays within x 90..1110."""
import os, subprocess, sys

W, H = 1200, 630
PAPER, PAPER2 = "#f6f8f6", "#e9eeeb"
INK, MUTED, RULE = "#15201c", "#55635e", "#c9d3ce"
TEAL, TEAL_INK, TEAL_TINT = "#00cb95", "#007558", "#d2f5e8"
AMBER, AMBER_TINT = "#a8620b", "#f6e5cc"
VIOLET, VIOLET_TINT = "#5b4fc4", "#e7e4f8"
SANS = "Inter"
MONO = "DejaVu Sans Mono"


def page(body, kicker):
    return f'''<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{H}" viewBox="0 0 {W} {H}">
<rect width="{W}" height="{H}" fill="{PAPER}"/>
<text x="96" y="72" font-family="{MONO}" font-size="17" fill="{MUTED}" letter-spacing="1">6G-RUPA <tspan fill="{RULE}">/</tspan> {kicker}</text>
{body}
</svg>'''


def t(x, y, s, size=20, fill=INK, anchor="start", weight=500, family=SANS, extra=""):
    return (f'<text x="{x}" y="{y}" font-family="{family}" font-size="{size}" fill="{fill}" '
            f'text-anchor="{anchor}" font-weight="{weight}" {extra}>{s}</text>')


# --- 5G core grows with its users: the chart ---------------------------------
def core_grows():
    x0, x1, yb, yt = 150, 880, 530, 150
    ymax = 500
    def Y(mb): return yb - mb / ymax * (yb - yt)
    def X(m): return x0 + m / 40 * (x1 - x0)
    b = []
    for mb in (100, 200, 300, 400, 500):
        b.append(f'<line x1="{x0}" x2="{x1}" y1="{Y(mb)}" y2="{Y(mb)}" stroke="{RULE}" stroke-width="1"/>')
        b.append(t(x0 - 14, Y(mb) + 6, f"{mb}", 16, MUTED, "end", 400, MONO))
    b.append(t(x0 - 14, yb + 6, "0", 16, MUTED, "end", 400, MONO))
    for m in (0, 10, 20, 30, 40):
        b.append(t(X(m), yb + 32, f"{m}M" if m else "0", 16, MUTED, "middle", 400, MONO))
    b.append(f'<path d="M{x0} {yt-10} V{yb} H{x1+10}" fill="none" stroke="{INK}" stroke-width="2"/>')
    b.append(t(x0, yt - 26, "forwarding state, MB", 17, MUTED, "start", 500))
    b.append(t(x1, yb + 66, "active sessions", 17, MUTED, "end", 500))
    # 5G: grows with every session
    b.append(f'<line x1="{X(0)}" y1="{Y(0)}" x2="{X(40)}" y2="{Y(460)}" stroke="{INK}" stroke-width="5" stroke-linecap="round"/>')
    b.append(f'<circle cx="{X(40)}" cy="{Y(460)}" r="7" fill="{INK}"/>')
    b.append(t(X(40) + 24, Y(460) - 4, "5G UPF", 26, INK, "start", 700))
    b.append(t(X(40) + 24, Y(460) + 26, "≈460 MB", 20, MUTED, "start", 400, MONO))
    # 6G-RUPA: flat
    b.append(f'<line x1="{X(0)}" y1="{Y(0)-3}" x2="{X(40)}" y2="{Y(0)-3}" stroke="{TEAL}" stroke-width="7" stroke-linecap="round"/>')
    b.append(f'<circle cx="{X(40)}" cy="{Y(0)-3}" r="8" fill="{TEAL}"/>')
    b.append(t(X(40) + 24, Y(0) - 22, "6G-RUPA", 26, TEAL_INK, "start", 700))
    b.append(t(X(40) + 24, Y(0) + 8, "0 MB", 20, MUTED, "start", 400, MONO))
    return page("\n".join(b), "UPF memory vs. users, simulated Movistar Spain")


# --- 5G already has layers: split N layer vs one layer -----------------------
def has_layers():
    cols = {"UE": 280, "gNB": 620, "UPF": 960}
    rows = {"5G": 250, "6G-RUPA": 430}
    bh = 64
    b = []
    for name, x in cols.items():
        b.append(f'<line x1="{x}" x2="{x}" y1="160" y2="510" stroke="{RULE}" stroke-width="2" stroke-dasharray="2 8" stroke-linecap="round"/>')
    b.append(t(cols["UE"], 556, "UE", 24, INK, "middle", 700))
    b.append(t(cols["gNB"], 556, "gNB", 24, INK, "middle", 700))
    b.append(t(cols["UPF"], 556, "UPF", 24, INK, "middle", 700))
    b.append(t(cols["UPF"], 584, "GUPF in 6G-RUPA", 16, MUTED, "middle", 400))
    for name, y in rows.items():
        b.append(t(96, y + 8, name, 24, TEAL_INK if name == "6G-RUPA" else INK, "start", 700))
    # 5G: two protocols meeting at the gNB
    y = rows["5G"] - bh / 2
    gap = 14
    b.append(f'<rect x="{cols["UE"]}" y="{y}" width="{cols["gNB"]-cols["UE"]-gap}" height="{bh}" rx="8" fill="{AMBER_TINT}" stroke="{AMBER}" stroke-width="2"/>')
    b.append(f'<rect x="{cols["gNB"]+gap}" y="{y}" width="{cols["UPF"]-cols["gNB"]-gap}" height="{bh}" rx="8" fill="{PAPER2}" stroke="{MUTED}" stroke-width="2"/>')
    b.append(t((cols["UE"] + cols["gNB"]) / 2, rows["5G"] + 8, "SDAP", 24, AMBER, "middle", 600, MONO))
    b.append(t((cols["gNB"] + cols["UPF"]) / 2, rows["5G"] + 8, "GTP-U", 24, INK, "middle", 600, MONO))
    # the translation point
    b.append(f'<line x1="{cols["gNB"]}" x2="{cols["gNB"]}" y1="{rows["5G"]-bh/2-6}" y2="{rows["5G"]+bh/2+6}" stroke="{INK}" stroke-width="3" stroke-linecap="round"/>')
    b.append(t(cols["gNB"], rows["5G"] - bh / 2 - 18, "translated here", 17, MUTED, "middle", 400))
    # 6G-RUPA: one layer
    y = rows["6G-RUPA"] - bh / 2
    b.append(f'<rect x="{cols["UE"]}" y="{y}" width="{cols["UPF"]-cols["UE"]}" height="{bh}" rx="8" fill="{TEAL_TINT}" stroke="{TEAL_INK}" stroke-width="2"/>')
    b.append(t((cols["UE"] + cols["UPF"]) / 2, rows["6G-RUPA"] + 8, "EFCP, end to end", 24, TEAL_INK, "middle", 600, MONO))
    return page("\n".join(b), "the user plane, read as layers")


# --- Follow a ping through the layers ----------------------------------------
def ping():
    cols = [("UE", 300, {"IP", "Operator A", "link"}),
            ("gNB", 515, {"Operator A", "link"}),
            ("I-GUPF", 730, {"Operator A", "link"}),
            ("GUPF", 945, {"IP", "Operator A", "link"})]
    bands = {"IP": 220, "Operator A": 330, "link": 440}
    cw, ch = 132, 58
    b = []
    for name, y in bands.items():
        b.append(f'<rect x="{200}" y="{y-ch/2-10}" width="{850}" height="{ch+20}" rx="10" fill="{TEAL_TINT if name == "Operator A" else PAPER2}" opacity="{0.55 if name == "Operator A" else 0.6}"/>')
        b.append(t(96, y + 6, name, 17, TEAL_INK if name == "Operator A" else MUTED, "start", 600))
    for name, x, has in cols:
        b.append(f'<rect x="{x-cw/2-14}" y="160" width="{cw+28}" height="345" rx="16" fill="none" stroke="{MUTED}" stroke-width="1.6"/>')
        b.append(t(x, 548, name, 24, INK, "middle", 700))
        for band, y in bands.items():
            if band in has:
                op = band == "Operator A"
                b.append(f'<rect x="{x-cw/2}" y="{y-ch/2}" width="{cw}" height="{ch}" rx="8" fill="{PAPER}" stroke="{TEAL_INK if op else INK}" stroke-width="{2.2 if op else 1.6}"/>')
    # the packet: down the UE, along links, up only as far as each box relays
    o = 14
    pts = [(300 - o, 220), (300 - o, 440), (515 - o, 440), (515 - o, 330), (515 + o, 330), (515 + o, 440),
           (730 - o, 440), (730 - o, 330), (730 + o, 330), (730 + o, 440), (945 - o, 440), (945 - o, 220), (1062, 220)]
    d = "M" + " L".join(f"{x} {y}" for x, y in pts)
    b.append(f'<path d="{d}" fill="none" stroke="{PAPER}" stroke-width="13" stroke-linejoin="round" stroke-linecap="round"/>')
    b.append(f'<path d="{d}" fill="none" stroke="{INK}" stroke-width="5" stroke-linejoin="round" stroke-linecap="round"/>')
    b.append(f'<circle cx="{300-o}" cy="220" r="9" fill="{INK}"/>')
    for x in (515, 730):
        b.append(f'<circle cx="{x}" cy="330" r="9" fill="{TEAL}" stroke="{PAPER}" stroke-width="3"/>')
    b.append(f'<circle cx="{945-o}" cy="220" r="9" fill="{INK}"/>')
    b.append(f'<path d="M1052 210 l14 10 l-14 10" fill="none" stroke="{INK}" stroke-width="5" stroke-linecap="round" stroke-linejoin="round"/>')
    b.append(t(1080, 148, "8.8.8.8", 20, INK, "end", 600, MONO))
    b.append(f'<circle cx="{96+8}" cy="596" r="8" fill="{TEAL}"/>')
    b.append(t(122, 602, "relays here, IP stays sealed", 17, MUTED, "start", 400))
    return page("\n".join(b), "ping 8.8.8.8, box by box")


# --- Network slicing from the UPF perspective --------------------------------
def slicing():
    flows = [("NSI 1", 220, TEAL_INK, TEAL_TINT, True),
             ("NSI 2", 320, VIOLET, VIOLET_TINT, True),
             ("NSI 3", 420, AMBER, AMBER_TINT, False)]
    gx0, gx1 = 500, 640
    b = []
    for label, y, c, tint, ok in flows:
        b.append(f'<rect x="96" y="{y-24}" width="120" height="48" rx="24" fill="{tint}" stroke="{c}" stroke-width="2"/>')
        b.append(t(156, y + 7, label, 20, c, "middle", 600, MONO))
        end = gx0 if ok else gx0 - 80
        b.append(f'<line x1="216" y1="{y}" x2="{end}" y2="{y}" stroke="{c}" stroke-width="5" stroke-linecap="round" {"" if ok else "stroke-dasharray=\"2 12\""}/>')
    # the allocator: where slicing actually lives
    b.append(f'<rect x="{gx0}" y="160" width="{gx1-gx0}" height="320" rx="14" fill="{PAPER2}" stroke="{INK}" stroke-width="2"/>')
    b.append(t((gx0 + gx1) / 2, 196, "flow", 19, INK, "middle", 700))
    b.append(t((gx0 + gx1) / 2, 220, "allocator", 19, INK, "middle", 700))
    b.append(t((gx0 + gx1) / 2, 330, "policy", 17, MUTED, "middle", 400))
    b.append(t((gx0 + gx1) / 2, 352, "caps · QoS", 17, MUTED, "middle", 400))
    # refused
    b.append(f'<path d="M{gx0-70} 410 l-18 20 m0 -20 l18 20" stroke="{AMBER}" stroke-width="4" stroke-linecap="round"/>')
    b.append(t(gx0 - 79, 470, "cap reached", 17, AMBER, "middle", 500))
    # admitted flows merge into one entry
    fx = 840
    b.append(f'<path d="M{gx1} 220 C {gx1+130} 220, {gx1+100} 270, {fx-40} 270" fill="none" stroke="{TEAL_INK}" stroke-width="5" stroke-linecap="round"/>')
    b.append(f'<path d="M{gx1} 320 C {gx1+130} 320, {gx1+100} 270, {fx-40} 270" fill="none" stroke="{VIOLET}" stroke-width="5" stroke-linecap="round"/>')
    b.append(f'<line x1="{fx-40}" y1="270" x2="{fx}" y2="270" stroke="{INK}" stroke-width="7" stroke-linecap="round"/>')
    # forwarding table
    tx, ty, tw = fx, 182, 270
    b.append(f'<rect x="{tx}" y="{ty}" width="{tw}" height="168" rx="12" fill="{PAPER}" stroke="{INK}" stroke-width="2"/>')
    b.append(t(tx + 20, ty + 36, "forwarding table", 18, MUTED, "start", 600))
    b.append(f'<line x1="{tx}" x2="{tx+tw}" y1="{ty+54}" y2="{ty+54}" stroke="{RULE}" stroke-width="1.5"/>')
    b.append(f'<rect x="{tx+12}" y="{ty+66}" width="{tw-24}" height="44" rx="6" fill="{PAPER2}"/>')
    b.append(t(tx + 26, ty + 96, "1.0  → core 2", 21, INK, "start", 600, MONO))
    b.append(t(tx + 20, ty + 146, "no slice IDs here", 17, MUTED, "start", 400))
    return page("\n".join(b), "a slice is policy at the allocator, not a forwarding entry")


THUMBS = {
    "content/posts/2026/5g-core-grows-with-its-users": core_grows,
    "content/posts/2026/5g-already-has-layers": has_layers,
    "content/posts/2026/follow-a-ping-through-the-layers": ping,
    "content/posts/2024/understanding-network-slicing": slicing,
}

# Usage: python3 scripts/6grupa-thumbnails.py . <svg-out-dir>   (needs rsvg-convert, Inter)
if __name__ == "__main__":
    site, out = sys.argv[1], sys.argv[2]
    os.makedirs(out, exist_ok=True)
    for bundle, fn in THUMBS.items():
        name = os.path.basename(bundle)
        svg = os.path.join(out, name + ".svg")
        open(svg, "w").write(fn())
        png = os.path.join(site, bundle, "featured.png")
        subprocess.run(["rsvg-convert", "-w", str(W * 2), "-h", str(H * 2), "-o", png, svg], check=True)
        print(png)
