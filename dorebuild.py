import subprocess
import os
import sys
from configparser import ConfigParser

component = ""
components = []
ldFlags = ""
cmd_dir = "cmd"
mainfile = ""

default_flags = {"BUILD_EXE": 1, "BUILD_RES": 1}

build_options = {
    "exe": {"BUILD_EXE": 1},
    "res": {"BUILD_RES": 1, "BUILD_RES_ECHO": 1},
    "all": default_flags,
    "": default_flags
}
build_flags = {
    "BUILD_RES_ECHO": 0,
    "BUILD_RES": 0,
    "BUILD_EXE": 0
}

def get_components():
    return [
        name for name in os.listdir(cmd_dir) 
        if os.path.isdir(os.path.join(cmd_dir, name))
    ]

def push_ldflag(v: str):
    global ldFlags
    if ldFlags=="":
        ldFlags=v
    else:
        ldFlags+= " " + v

def handle_cfg():
    global mainfile, ldFlags
    ldFlags = ""
    mainfile = "main"
    if not os.path.isfile(f"cmd/{component}/build.cfg"):
        return
    
    cfg = ConfigParser(allow_no_value=True)
    try: 
        cfg.read(f"cmd/{component}/build.cfg")
    except Exception:
        return
    
    mf = cfg.get('app', 'main')
    if mf != "":
        mainfile = mf

    header = cfg.get('app', 'header')
    if header=="" and cfg.getboolean('app', 'gui', fallback=False):
        header="-H windowsgui"
    else:
        header=f"-H {header}"

    push_ldflag(header)
    if cfg.get('app', "useBuild"):
        push_ldflag(f"-X '{mainfile}.Build=true'")

    variables = cfg.items("variables")
    for k, v in variables:
        push_ldflag(f"-X '{k}={v}'")

def lookup_flags():
    global components, build_options, build_flags
    
    if len(sys.argv) >= 2:
        tp = sys.argv[1]
        if tp in build_options:
            build_flags.update(build_options[tp])
        else:
            print(f"Error: Unexpected first argument value {tp} (only \"all\", \"exe\", \"res\" - allowed)")
            sys.exit(1)
    else:
        build_flags.update(default_flags)
   
    if len(sys.argv) > 2:
        component = sys.argv[2]
        if component == "*":  # Including all components
            components = get_components()
        else:
            components = [component]
    else:
        components = get_components()
    
def build_res():
    if not os.path.isdir(f"assets/{component}"):
        if build_flags["BUILD_RES_ECHO"]:
            print(f"Warning: Folder assets/{component} not found, nothing to build", file=sys.stderr)
        return
    tpath = f"cmd/{component}/resources"
    if not os.path.isdir(tpath):
        if build_flags["BUILD_RES_ECHO"]:
            print(f"Error: Component internal package 'resources' not found (cmd/{component}/resources), no target to build resources", file=sys.stderr)
        return
    windres = [
        "windres",
        "-i",
        "resources.rc",
        "-o",
        f"../../cmd/{component}/resources/resources.syso",
    ]
    try:
        subprocess.run(windres, check=True, capture_output=True, text=True, cwd=f"assets/{component}")
    except subprocess.CalledProcessError as e:
        print("Error during 'resources.syso' build")
        print(e.stderr)
        sys.exit(e.returncode)
    gcc = [
        "gcc",
        "-shared",
        "-o",
        f"debug/{component}Res.dll",
        f"cmd/{component}/resources/resources.syso",
    ]
    try: 
        subprocess.run(gcc, check=True, capture_output=True, text=True)
    except subprocess.CalledProcessError as e:
        print(f"Error during {component}Res.dll build")
        print(e.stderr)
        sys.exit(e.returncode)

def build_exe():
    if not os.path.exists("bin"): 
        os.mkdir("bin")
    else: 
        if os.path.exists(f"bin/{component}.exe"):
            os.remove(f"bin/{component}.exe")
    go = [
        "go", "build", "-ldflags", ldFlags,
        "-o", f"bin/{component}.exe",
        f"cmd/{component}/{mainfile}.go",
    ]
    go_env = os.environ.copy()
    go_env["GOOS"]="windows"
    go_env["GOARCH"]="amd64"
    try:
        subprocess.run(go, check=True, capture_output=True, text=True, env=go_env)

    except subprocess.CalledProcessError as e:
        print("Error during build go file ")
        print(e.stderr)
        sys.exit(e.returncode)

def main(): 
    global components, component
    lookup_flags()
    for component in components:
        if build_flags["BUILD_RES"]: 
            build_res()
        if build_flags["BUILD_EXE"]:
            build_exe()

if __name__ == "__main__":
    main()
