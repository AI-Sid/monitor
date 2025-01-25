import subprocess
import os
import sys

component = ""
ldFlags = ""

default_flags = {"BUILD_EXE": 1, "BUILD_RES": 1}

build_options = {
    "con": {"BUILD_CON": 1, "BUILD_EXE": 1},
    "exe": {"BUILD_EXE": 1},
    "res": {"BUILD_RES": 1, "BUILD_RES_ECHO": 1},
    "all": default_flags,
    "": default_flags,
}
build_flags = {
    "BUILD_RES_ECHO": 0,
    "BUILD_RES": 0,
    "BUILD_EXE": 0,
    "BUILD_CON": 0,
}

def lookup_flags():
    global component, build_options, build_flags, ldFlags
    if len(sys.argv) < 2:
        print("Error: Specify a component name as the first parameter")
        sys.exit(1)

    component = sys.argv[1]
    if not os.path.isdir(f"cmd/{component}"):
        print(f"Error: Component {component} not found (folder cmd/{component} is not found)")
        sys.exit(1)
    
    if len(sys.argv) > 2:
        tp = sys.argv[2]
        if tp in build_options:
            build_flags.update(build_options[tp])
        else:
            print(f"Error: Unexpected second parameter with name {tp} (only \"all\", \"exe\", \"con\", \"res\", \"\" - allowed)")
            sys.exit(1)
    else:
        build_flags.update(default_flags)
  
    extra_ldflags = ""

    if len(sys.argv) > 3:
        extra_ldflags = sys.argv[3]

    if not build_flags["BUILD_CON"]:
        ldFlags = "" #"-H windowsgui"
        if extra_ldflags != "":
            ldFlags += " " + extra_ldflags
    else:
        ldFlags = extra_ldflags

def build_res():
    if not os.path.isdir(f"assets/{component}"):
        if build_flags["BUILD_RES_ECHO"]:
            print(f"Warning: Folder assets/{component} not found, Nothing to build")
        return
    tpath = f"cmd/{component}/resources"
    if not os.path.isdir(tpath):
        print(f"Error: Component internal package resources not found (cmd/{component}/resources)")
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
        print("Error during resources.syso build")
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
        f"cmd/{component}/main.go",
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
    lookup_flags()
    if build_flags["BUILD_RES"]: 
        build_res()
    if build_flags["BUILD_EXE"]:
        build_exe()

if __name__ == "__main__":
    main()
