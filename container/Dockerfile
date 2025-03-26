# BUILDER
FROM quay.io/centos/centos:stream9 AS builder

# Add ARGs to enable or disable the Demo config
# Set this argument to any value to enable the Demo config
# Demo mode is enabled by appending --build_arg _EN_DEMO=true to the build command
# Demo mode is disabled by ommiting the --build_arg option
ARG _EN_DEMO
RUN if [[ -z "$_EN_DEMO" ]] ; \
	then \
		echo "Demo is disabled" ; \
	else \
		echo "Demo is enabled" ;\
	fi

# build ARGs
ARG _RADARE2_GIT_REPO="https://github.com/radareorg/radare2.git"
ENV RADARE2_GIT_REPO=$_RADARE2_GIT_REPO
# Radare2 constitutes a noteworthy dependency. 
# However, it's a software that undergoes consistent changes across commits. 
# In my experience, even within the same minor version, significant alterations can be observed. 
# Despite ks-nav not utilizing any uncommon features of Radare2, I prefer to adhere to the most thoroughly tested commit.
ARG _RADARE2_GIT_COMMIT=4a07098809c71c21515d0f862ac4ca044ae95e95
ENV RADARE2_GIT_COMMIT=$_RADARE2_GIT_COMMIT

ARG _KS_NAV_GIT_REPO="https://github.com/elisa-tech/ks-nav"
ENV KS_NAV_GIT_REPO=$_KS_NAV_GIT_REPO
ARG _KS_NAV_GIT_REPO_BRANCH=main
ENV KS_NAV_GIT_REPO_BRANCH=$_KS_NAV_GIT_REPO_BRANCH

# install build dependecies
RUN dnf install -y autoconf automake gcc gcc-c++ git make patch
RUN dnf install -y golang
RUN dnf install -y kernel-devel

# build app
WORKDIR /build

# Build radare2
RUN git clone ${RADARE2_GIT_REPO} /build/radare2 && cd /build/radare2 && git checkout ${RADARE2_GIT_COMMIT}
RUN cd /build/radare2 && ./configure && make && make install

# Build ks-nav
RUN git clone ${KS_NAV_GIT_REPO} /build/ksnav && cd /build/ksnav/kern_bin_db && git checkout ${KS_NAV_GIT_REPO_BRANCH}
# This local file is copied into the build folder and used for the demo (to support a local database)
COPY ksnav_nav_local.json /build/.
RUN if [[-z "$_EN_DEMO" ]] ; \
	then \
		echo "Demo not running" ; \
	else \
		cp /build/ksnav_nav_local.json /build/ksnav/navweb/data/configs/container.json ; \
	fi

RUN cd /build/ksnav/kern_bin_db && make && make install
RUN cd /build/ksnav/nav && make && make install
RUN cd /build/ksnav/navweb && go get -u github.com/go-bindata/go-bindata/... && go install github.com/go-bindata/go-bindata/... && make && make install


##############################################################################################################################################

# RUNTIME
FROM quay.io/centos/centos:stream9 

# run ARGs
ARG _POSTGRES_NAME=only
ENV POSTGRES_NAME=$_POSTGRES_NAME
#ARG _POSTGRES_DATA_DIR=/var/lib/postgresql/data/
ARG _POSTGRES_DATA_DIR=/var/lib/pgsql/data/
ENV POSTGRES_DATA_DIR=$_POSTGRES_DATA_DIR
ARG _POSTGRES_IDENT_FILE=pg_ident.conf
ENV POSTGRES_IDENT_FILE=$_POSTGRES_IDENT_FILE
ARG _POSTGRES_CONF_FILE=${POSTGRES_DATA_DIR}/${POSTGRES_NAME}/postgresql.conf
ENV POSTGRES_CONF_FILE=$_POSTGRES_CONF_FILE

ARG _START_SCRIPT_NAME=/usr/bin/start.sh
ENV START_SCRIPT_NAME=$_START_SCRIPT_NAME

# install dependencies
RUN dnf module -y install postgresql:15
RUN dnf install -y binutils
RUN dnf install -y graphviz

# setup resources

WORKDIR /app
COPY postgres_conf_template.conf /tmp
# pg.db.tar.gz is a link, please set it to empty database, or whatever db you have.
COPY pg.db.tar.gz /tmp
RUN chown -R postgres:postgres ${POSTGRES_DATA_DIR} && \
	chmod 750 ${POSTGRES_DATA_DIR}
RUN [ ! -d ${POSTGRES_DATA_DIR}/${POSTGRES_NAME} ] && \
	mkdir -p ${POSTGRES_DATA_DIR}/${POSTGRES_NAME} && \
	cd ${POSTGRES_DATA_DIR}/${POSTGRES_NAME} && \
	tar -zxovf /tmp/pg.db.tar.gz && \
	chown -R postgres:postgres ${POSTGRES_DATA_DIR}/${POSTGRES_NAME} && \
	chmod 750 ${POSTGRES_DATA_DIR}/${POSTGRES_NAME}  && \
	cd - && echo DONE
RUN cat /tmp/postgres_conf_template.conf | \
	sed -r "s|%POSTGRES_DATA_DIR%|${POSTGRES_DATA_DIR}|g" | \
	sed -r "s|%POSTGRES_NAME%|${POSTGRES_NAME}|g" | \
	sed -r "s|%POSTGRES_DATA_DIR%|${POSTGRES_DATA_DIR}|g" | \
	sed -r "s|%POSTGRES_CONF_FILE%|${POSTGRES_CONF_FILE}|g" | \
	sed -r "s|%POSTGRES_IDENT_FILE%|${POSTGRES_IDENT_FILE}|g" | tee ${POSTGRES_CONF_FILE}
RUN sed -i /lc_/d ${POSTGRES_CONF_FILE}
RUN mkdir -p /run/postgresql && chown postgres:postgres /run/postgresql
COPY --from=builder /usr/bin/nav-db-filler /usr/bin
COPY --from=builder /usr/bin/nav /usr/bin
COPY --from=builder /usr/bin/navweb /usr/bin
COPY --from=builder /build/ksnav/kern_bin_db/psql.sql /tmp
COPY --from=builder /build/radare2/libr/reg/libr_reg.so /usr/lib64
COPY --from=builder /build/radare2/libr/syscall/libr_syscall.so /usr/lib64
COPY --from=builder /build/radare2/libr/lang/libr_lang.so /usr/lib64
COPY --from=builder /build/radare2/libr/main/libr_main.so /usr/lib64
COPY --from=builder /build/radare2/libr/io/libr_io.so /usr/lib64
COPY --from=builder /build/radare2/libr/core/libr_core.so /usr/lib64
COPY --from=builder /build/radare2/libr/fs/libr_fs.so /usr/lib64
COPY --from=builder /build/radare2/libr/util/libr_util.so /usr/lib64
COPY --from=builder /build/radare2/libr/magic/libr_magic.so /usr/lib64
COPY --from=builder /build/radare2/libr/bin/libr_bin.so /usr/lib64
COPY --from=builder /build/radare2/libr/debug/libr_debug.so /usr/lib64
COPY --from=builder /build/radare2/libr/cons/libr_cons.so /usr/lib64
COPY --from=builder /build/radare2/libr/flag/libr_flag.so /usr/lib64
COPY --from=builder /build/radare2/libr/esil/libr_esil.so /usr/lib64
COPY --from=builder /build/radare2/libr/config/libr_config.so /usr/lib64
COPY --from=builder /build/radare2/libr/egg/libr_egg.so /usr/lib64
COPY --from=builder /build/radare2/libr/asm/libr_asm.so /usr/lib64
COPY --from=builder /build/radare2/libr/socket/libr_socket.so /usr/lib64
COPY --from=builder /build/radare2/libr/bp/libr_bp.so /usr/lib64
COPY --from=builder /build/radare2/libr/search/libr_search.so /usr/lib64
COPY --from=builder /build/radare2/libr/arch/libr_arch.so /usr/lib64
COPY --from=builder /build/radare2/libr/crypto/libr_crypto.so /usr/lib64
COPY --from=builder /build/radare2/libr/anal/libr_anal.so /usr/lib64
COPY --from=builder /build/radare2/binr/radare2/radare2 /usr/bin
EXPOSE 8080

# start the thing
RUN find /app
RUN echo -e '#!/bin/sh\nchown -R postgres:postgres ${POSTGRES_DATA_DIR} /run/postgresql\nsu postgres -c "pg_ctl start -D ${POSTGRES_DATA_DIR}/${POSTGRES_NAME}"\n/usr/bin/navweb\n' >/usr/bin/start.sh && chmod +x /usr/bin/start.sh
ENTRYPOINT ["/usr/bin/start.sh"]


