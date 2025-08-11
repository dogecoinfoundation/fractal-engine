


dogenet genkey dev-key
dogenet genkey ident-key ident-pub

CMD export KEY=$(cat dev-key) && export IDENT=$(cat ident-pub) && ./dogenet --local --public 127.0.0.1