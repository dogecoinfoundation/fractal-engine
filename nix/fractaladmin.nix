{ lib, buildNpmPackage, fetchFromGitHub, nodejs, prisma-engines }:

buildNpmPackage rec {
  pname = "fractaladmin";
  version = "3c156b4";

  src = fetchFromGitHub {
    owner = "dogecoinfoundation";
    repo = "fractal-ui";
    rev = "3c156b416acef7f46530928fab1224cce3b624a2";
    sha256 = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # TODO: Add correct hash
  };

  npmDepsHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # TODO: Add correct hash

  # Ensure Prisma can find its engines
  PRISMA_QUERY_ENGINE_LIBRARY = "${prisma-engines}/lib/libquery_engine.node";
  PRISMA_QUERY_ENGINE_BINARY = "${prisma-engines}/bin/query-engine";
  PRISMA_SCHEMA_ENGINE_BINARY = "${prisma-engines}/bin/schema-engine";

  preBuild = ''
    # Set up data directory
    mkdir -p data
    chmod 0777 data
    
    # Generate Prisma client and run migrations
    export DATABASE_URL=''${DATABASE_URL:-"file:./data/dev.db"}
    npx prisma generate
    npx prisma migrate deploy || echo "Migration failed, continuing..."
  '';

  buildPhase = ''
    runHook preBuild
    npm run build
    runHook postBuild
  '';

  installPhase = ''
    runHook preInstall
    
    mkdir -p $out/lib/fractaladmin
    cp -r . $out/lib/fractaladmin/
    
    # Create wrapper script
    mkdir -p $out/bin
    cat > $out/bin/fractaladmin << 'EOF'
    #!/usr/bin/env bash
    
    export NODE_ENV=production
    export DATABASE_URL=''${DATABASE_URL:-"file://$HOME/.fractaladmin/dev.db"}
    
    # Ensure data directory exists
    mkdir -p $(dirname $(echo $DATABASE_URL | sed 's/file://'))
    
    cd $out/lib/fractaladmin
    
    # Run migrations if needed
    ${nodejs}/bin/npx prisma migrate deploy || true
    
    # Start the application
    exec ${nodejs}/bin/npm start
    EOF
    
    chmod +x $out/bin/fractaladmin
    
    runHook postInstall
  '';

  meta = with lib; {
    description = "Fractal UI Admin interface";
    homepage = "https://github.com/dogecoinfoundation/fractal-ui";
    license = licenses.mit;
    platforms = platforms.all;
  };
}
