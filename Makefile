.PHONY: all res exe

comp := *
# Общая цель для всех компонентов
all: 
	@python3 dorebuild.py all $(comp)
res:
	@python3 dorebuild.py res $(comp)
exe: 
	@python3 dorebuild.py exe $(comp)
