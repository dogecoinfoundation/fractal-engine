{ lib, buildNpmPackage, fetchFromGitHub, nodejs, cacert, pkgs }:

buildNpmPackage rec {
  pname = "fractaladmin";
  version = "main";

  src = fetchFromGitHub {
    owner = "dogecoinfoundation";
    repo = "fractal-ui";
    rev = "main";
    sha256 = "sha256-BmK+p9ovjtoGoN/VHI1QuqYwm19DUsClSOAyDX/Rj1Y=";
  };

  postPatch = ''
    cp ${./package-lock.json} package-lock.json
  '';

  npmDepsHash = "sha256-HJWjb1RbrGijyCo88Nf4lMI5I331bvF6+OLjRPU0kMI=";

  nativeBuildInputs = [ cacert pkgs.prisma-engines ];

  preBuild = ''
    export DATABASE_URL="file:./dev.db"
    export NEXT_TELEMETRY_DISABLED=1
    export PRISMA_QUERY_ENGINE_LIBRARY=${pkgs.prisma-engines}/lib/libquery_engine.node
    export PRISMA_QUERY_ENGINE_BINARY=${pkgs.prisma-engines}/bin/query-engine
    export PRISMA_SCHEMA_ENGINE_BINARY=${pkgs.prisma-engines}/bin/schema-engine
    export PRISMA_INTROSPECTION_ENGINE_BINARY=${pkgs.prisma-engines}/bin/introspection-engine
    export PRISMA_FMT_BINARY=${pkgs.prisma-engines}/bin/prisma-fmt
    
    # Generate Prisma client and run migrations
    npx prisma generate
    npx prisma migrate deploy
  '';

  npmBuildScript = "build";

  installPhase = ''
    mkdir -p $out/lib/fractaladmin
    cp -r .next/ $out/lib/fractaladmin/
    cp -r public/ $out/lib/fractaladmin/ || true
    cp package.json $out/lib/fractaladmin/
    cp next.config.ts $out/lib/fractaladmin/ || true
    cp -r node_modules/ $out/lib/fractaladmin/
    cp -r generated/ $out/lib/fractaladmin/ || true
    cp -r prisma/ $out/lib/fractaladmin/ || true

    # Copy Prisma engines
    mkdir -p $out/lib/fractaladmin/.prisma/client
    cp ${pkgs.prisma-engines}/lib/libquery_engine.node $out/lib/fractaladmin/.prisma/client/ || true

    # Create wrapper script
    mkdir -p $out/bin
    cat > $out/bin/fractaladmin << EOF
#!/usr/bin/env bash
export PRISMA_QUERY_ENGINE_LIBRARY=$out/lib/fractaladmin/.prisma/client/libquery_engine.node
export PRISMA_QUERY_ENGINE_BINARY=${pkgs.prisma-engines}/bin/query-engine
export DATABASE_URL=\''${DATABASE_URL:-file:./dev.db}
cd $out/lib/fractaladmin
exec ${nodejs}/bin/node node_modules/next/dist/bin/next start "\$@"
EOF
    chmod +x $out/bin/fractaladmin
  '';

  meta = with lib; {
    description = "Fractal Admin UI - Web interface for Dogecoin fractal services";
    homepage = "https://github.com/dogecoinfoundation/fractal-ui";
    license = licenses.mit;
    maintainers = [ ];
    platforms = platforms.linux ++ platforms.darwin;
  };
}
