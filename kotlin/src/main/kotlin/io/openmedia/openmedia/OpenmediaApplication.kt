package io.openmedia.openmedia

import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication

@SpringBootApplication
class OpenMediaApplication

fun main(args: Array<String>) {
  runApplication<OpenMediaApplication>(*args)
}
