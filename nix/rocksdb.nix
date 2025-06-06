{ lib
, stdenv
, fetchFromGitHub
, fetchpatch
, cmake
, ninja
, bzip2
, lz4
, snappy
, zlib
, zstd
, windows
, enableLite ? false
, enableShared ? !stdenv.hostPlatform.isStatic
, sse42Support ? stdenv.hostPlatform.sse4_2Support
}:

stdenv.mkDerivation rec {
  pname = "rocksdb";
  version = "9.8.4";

  src = fetchFromGitHub {
    owner = "facebook";
    repo = pname;
    rev = "v${version}";
    sha256 = "sha256-A6Gx4FqoGlxITUUz9k6tkDjUcLtMUBK9JS8vuAS96H0=";
  };

  nativeBuildInputs = [ cmake ninja ];

  propagatedBuildInputs = [ bzip2 lz4 snappy zlib zstd ];

  outputs = [
    "out"
    "tools"
  ];

  NIX_CFLAGS_COMPILE = lib.optionals stdenv.cc.isGNU [
    "-Wno-error=deprecated-copy"
    "-Wno-error=pessimizing-move"
    # Needed with GCC 12
    "-Wno-error=format-truncation"
    "-Wno-error=maybe-uninitialized"
  ] ++ lib.optionals stdenv.cc.isClang [
    "-Wno-error=unused-private-field"
    "-faligned-allocation"
  ];

  cmakeFlags = [
    "-DPORTABLE=1"
    "-DWITH_JNI=0"
    "-DWITH_BENCHMARK_TOOLS=0"
    "-DWITH_TESTS=1"
    "-DWITH_TOOLS=0"
    "-DWITH_CORE_TOOLS=1"
    "-DWITH_BZ2=1"
    "-DWITH_LZ4=1"
    "-DWITH_SNAPPY=1"
    "-DWITH_ZLIB=1"
    "-DWITH_ZSTD=1"
    "-DWITH_GFLAGS=0"
    "-DUSE_RTTI=1"
    (lib.optional sse42Support "-DFORCE_SSE42=1")
    (lib.optional enableLite "-DROCKSDB_LITE=1")
    "-DFAIL_ON_WARNINGS=${if stdenv.hostPlatform.isMinGW then "NO" else "YES"}"
  ] ++ lib.optional (!enableShared) "-DROCKSDB_BUILD_SHARED=0";

  # otherwise "cc1: error: -Wformat-security ignored without -Wformat [-Werror=format-security]"
  hardeningDisable = lib.optional stdenv.hostPlatform.isWindows "format";

  preInstall = ''
    mkdir -p $tools/bin
    cp tools/{ldb,sst_dump}${stdenv.hostPlatform.extensions.executable} $tools/bin/
  '' + lib.optionalString stdenv.isDarwin ''
    ls -1 $tools/bin/* | xargs -I{} ${stdenv.cc.bintools.targetPrefix}install_name_tool -change "@rpath/librocksdb.${lib.versions.major version}.dylib" $out/lib/librocksdb.dylib {}
  '' + lib.optionalString (stdenv.isLinux && enableShared) ''
    ls -1 $tools/bin/* | xargs -I{} patchelf --set-rpath $out/lib:${stdenv.cc.cc.lib}/lib {}
  '';

  # Old version doesn't ship the .pc file, new version puts wrong paths in there.
  postFixup = ''
    if [ -f "$out"/lib/pkgconfig/rocksdb.pc ]; then
      substituteInPlace "$out"/lib/pkgconfig/rocksdb.pc \
        --replace '="''${prefix}//' '="/'
    fi
  '' + lib.optionalString stdenv.isDarwin ''
    ${stdenv.cc.targetPrefix}install_name_tool -change "@rpath/libsnappy.1.dylib" "${snappy}/lib/libsnappy.1.dylib" $out/lib/librocksdb.dylib
    ${stdenv.cc.targetPrefix}install_name_tool -change "@rpath/librocksdb.${lib.versions.major version}.dylib" "$out/lib/librocksdb.${lib.versions.major version}.dylib" $out/lib/librocksdb.dylib
  '';

  meta = with lib; {
    homepage = "https://rocksdb.org";
    description = "A library that provides an embeddable, persistent key-value store for fast storage";
    changelog = "https://github.com/facebook/rocksdb/raw/v${version}/HISTORY.md";
    license = licenses.asl20;
    platforms = platforms.all;
    maintainers = with maintainers; [ adev magenbluten ];
  };
}
