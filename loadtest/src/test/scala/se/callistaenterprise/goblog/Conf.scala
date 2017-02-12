package se.callistaenterprise.goblog

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import io.gatling.jdbc.Predef._

object Conf {
	var users = System.getProperty("users", "2").toInt
	val baseUrl = System.getProperty("baseUrl", "http://localhost:6767")
	var httpConf = http.baseURL(baseUrl)
	var duration = System.getProperty("duration", "30").toInt
}