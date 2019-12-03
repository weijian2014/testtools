cmake_minimum_required(VERSION 2.8)
project(simulator)

#set(CMAKE_CXX_STANDARD 17)

add_executable(Simulator Simulator.cpp)

#可执行文件路径(PROJECT_SOURCE_DIR为项目的根目录 PROJECT_BINARY_DIR为执行cmake时的目录)
set(EXECUTABLE_OUTPUT_PATH "${PROJECT_BINARY_DIR}")