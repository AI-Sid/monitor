# Proxy Settings Monitor for Windows
## 1. Вступ
Програма виконує моніторінг налаштувань **Proxy**. Програма використвує наступні бібліотеки:
- **golang.org/x/sys/windows** (cтворення локальних *Events*, іменованих *Mutex*, *Events*, завантаження *DLL*, робота з ресурсами
- **golang.org/x/sys/windows/registry** для спостереження за змінами реєстру для *HKEY_CURRENT_USER\SOFTWARE\Microsoft\Windows\CurrentVersion\Internet Settings*)
- **github.com/getlantern/systray** (Створення та керування *Windows Tray Icon*)
Системні інструменти:
- MSYS2 (https://www.msys2.org/) з встановленними minGW-w64, та minGW-w64-x86_64-make. Шлях до бінарних файлів minGW64 (наприклад, ***C:\msys64\mingw64\bin***)повинен бути доданий до змінної середовища PATH. MinGW64 використовується для компіляції файлів ресурсів *.rc в файли *.syso, для створення ресурсних .dll (тека Debug) та виконання make. MSYS2 має вбудований Python, що використовується під час складання файлів ресурсів або програм.

Під час кожного старту моніторінгу - програма записує поточні дані в журнал (%APPDATA%/appname/appname.log), потім під час кожної зміни параметрів - доповнює його новими записами.

## 2. Шаблон проекту
```
-monitor
    ├─ assets
    │   ├─ <component>               // component resources
    │   │   ├─ <some resource files>
    │   │   └─ resources.rc          // file to compile component resources
    │  ...                           // assets for other components
    ├─ cmd
    │   ├─ <component>               // component declaration
    │   │   ├─ [resources]           // optional, only if program has resources
    │   │   │   ├─ resources.go      // dummy file
    │   │   │   └─ [resources.syso]  // builded from assets/<component>/resources.rc
    │   │   ├─ <main>.go             
    │   │   ├─ [build.cfg]           // python config file for component building
    │   │   └─ ...                   // other packages  
    │  ...                           // other components
    ├─ debug
    │   ├─ <component>Res.dll        // optional, resource dll-files for debugging
    │  ...                           // other resource dll-files files 
    ├─ internal                      // internal packages
    │   ├─ <package>                 // package template
    │   │   └─ <go-files>
    │  ...                           // other packages
    ├─ scripts
    │   ├─ doBuild.py                // Python script for components building (used in batch-files and Makefile)
    │   ├─ rebuild.bat               // Batch file for rebuilding one component
    │   └─ rebuildAll.bat            // Batch file for rebuilding all components
    ├─ dorebuild.py                  // Python script used for building components
    ├─ go.mod
    ├─ go.sum
    ├─ LICENSE
    ├─ Makefile                      // Makefile for rebuild
    ├─ README.md
    └─ rebuild.bat                   // Rebuild batch file for windows
```
## 2. Збірка компонентів

Кожен компонент може мати власний файл ***build.cfg***, який має структуру схожу на INI-файл (див. Python Config Parser).\
В цьому файлі сприймаються тільки дві секції: ***app*** та ***variables***.\
В секції ***app*** можуть бути наступні параметри:\
    - **header**. Встановлює значення *header* для **ldflags** (зазвичай, цей параметр відсутній, або має значення "windowsgui"/"windows"/etc.)\
    - **gui**. Якщо значення = *true*, **header** - відсутній, або має пусте значення, присвоює в **header** значення "windowsgui"\
    - **main**. Якщо значення не пусте, підмінює назву головного файлу компоненту на вказане в **main**. Значення параметру не повинно містити розширення ".go"\
    - **useBuild**. Якщо значення = *true*, додає до **ldflags** встановлення змінної **\<main\>**.Build=true
   
В секції ***variables** вказуються додаткові значення змінних, які можуть бути використані під час побудови виконавчого файлу. Назва - шлях до змінної файлу, значення - це те що повинно бути заміненим перед складанням бінарного файлу (не залежить від значення параметру **main**). Наприклад:\
    - myapp.Version=1.0.0    
    - myapp.Author=Unknown   

Приклад файлу ***build.cfg***

```
[app]
gui: true
useBuild: true

[variables]
main.Version: 1.0.0
```
Крім того, якщо компонент використовує ресурси, він повинен мати вкладену теку "resources", з пустим файлом-заглушкою, наприклад, resources.go:
```
package resources
```

### 2.1. Виконання скриптів для збірки компонентів

Для збірки використувуються наступні файли:   
- dorebuild.py - головна програма на Python   
- Makefile   
- rebuild.bat   
   
Makefile, rebuild.bat використовують dorebuild.py, та мають однакові аргументи:   
- **type**. Тип збірки. Може мати значення all (res + exe), res - збірка файлів ресурсів, exe - збірка бінарних файлів   
- **component**. Назва компоненту (тека *cmd/{component}*), або "*" (за замовченням) для всіх компонентів

Якщо жоден параметр не заданий - буде виконана збірка з аргументами: all, *

Для виконання Makefile можна використати **mingw32-make** з аргументами **type**, **component**, наприклад
```
mingw32-make all *
```
З консолі cmd/powershell (якщо make встановлено), або з консолі системи Linux (див. WSL):
```
make all *
```
Використати rebuild.bat з консолі cmd/powershell:
```
rebuild all *
```

Якщо тип збірки - res, або all, буде створено файл resources.syso в теці cmd/{component}/resources якщо така тека існує, та
якщо існує тека assets/{component} з файлом resources.rc.
Якщо тип збірки - exe, або all, буде створено новий бінарний файл bin/{component}.exe (Якщо тека **bin** не існує - вона буде створена)

Файли debug/*.dll використовуються для завантаження ресурсів під час виконання в debug-режимі, або через go run cmd/{component}/{main}.go.
Ці файли не мають відношення до бінарних файлів, останні їх не потребують.

## 3. Компонени

В репозіторії є тільки один компонент - ***proxyMon***, який створено відповідно завданню

## 4. Запуск ***proxyMon***

Компонент ***proxyMon*** може мати параметри (відповідно завданню: -start, -stop, -quit) \

Перший запуск програми (без -quit), вважається головною програмою. Всі інші запуски - тільки керують головною програмою. 

Програму можна запустити за допомогою **GO**
```
go run cmd/proxyMon/main.go <params>
```
В такому випадку можна спостерігати сповіщення в консолі

## 5. Примітки

1) В даному проекті відсутні тести.
2) Файл журналу збережено як вказано в завданні: %APPDATA%/appname/appname.log, тобто не %APPDATA%/proxyMon/proxyMon.log
3) Оскільки в завданні нічого не сказано про запис зміни статусу монітору, коли виконується start - відразу створюється запис у журналі, незалежно від того, змінювався він чи ні (ми не знаємо про можливі зміни proxy за час відсутності або простою монітору)
