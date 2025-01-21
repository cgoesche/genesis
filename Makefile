SHELL ::= /bin/bash
GIT_ROOT_DIR ::= $(shell git rev-parse --show-toplevel 2>/dev/null )

.SILENT: install-git-hooks
install-git-hooks:
	if [[ "${GIT_ROOT_DIR}" == "" ]]; then 																				\
		printf "[$$(tput setaf 1)$$(tput bold)error$$(tput sgr0)] %s\n" "Make sure you are in a git work tree"; 		\
		exit 1; 																								\
	fi
	rm -rf $(GIT_ROOT_DIR)/.git/hooks/*;
	printf "[$$(tput setaf 2)$$(tput bold)info$$(tput sgr0)] %s\n" "Copying git-hooks ..."; 								\
	for hook in $$(realpath -- ${GIT_ROOT_DIR}/githooks/???*); do														\
		chmod +x $${hook};																							\
		cp -a $${hook} ${GIT_ROOT_DIR}/.git/hooks;																		\
	done;																										\
	printf "[$$(tput setaf 2)$$(tput bold)success$$(tput sgr0)] %s\n" "git-hooks installed.";

.SILENT: install-pre-commit-config
install-pre-commit-config:
	if [[ "${GIT_ROOT_DIR}" == "" ]]; then 																				\
		printf "[$$(tput setaf 1)$$(tput bold)error$$(tput sgr0)] %s\n" "Make sure you are in a git work tree"; 		\
		exit 1; 																								\
	fi
	rm -rf $(GIT_ROOT_DIR)/.pre-commit-config.yaml;
	printf "[$$(tput setaf 2)$$(tput bold)info$$(tput sgr0)] %s\n" "Copying .pre-commit-config.yaml ..."; 								\
	cp -a $$(realpath -- ${GIT_ROOT_DIR}/pre-commit/.pre-commit-config.yaml) ${GIT_ROOT_DIR}/; \
	printf "[$$(tput setaf 2)$$(tput bold)success$$(tput sgr0)] %s\n" "pre-commit config installed."; \

.SILENT: devsetup
devsetup: install-git-hooks install-pre-commit-config
	printf "[$$(tput setaf 2)$$(tput bold)done$$(tput sgr0)] %s\n" "Development environment succesffully set up."; \
