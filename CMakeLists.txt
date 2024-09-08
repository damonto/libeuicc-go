cmake_minimum_required(VERSION 3.8)

project(libeuicc
    VERSION 2.1.0
    HOMEPAGE_URL "https://github.com/estkme-group/lpac"
    DESCRIPTION "C-based eUICC LPA."
    LANGUAGES C)

set(CMAKE_C_STANDARD 99)
set(CMAKE_C_STANDARD_REQUIRED ON)

set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/libs)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/libs)

if (APPLE)
    set(CMAKE_OSX_ARCHITECTURES "arm64;x86_64")
endif()

add_subdirectory(lpac/cjson)
add_subdirectory(lpac/euicc)
