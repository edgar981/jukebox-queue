plugins {
    kotlin("jvm") version "2.1.21"
    kotlin("plugin.serialization") version "2.1.21"
    application
}

group = "com.github.edgar981"
version = "0.1.0"

repositories { mavenCentral() }

val ktor = "3.1.3"

dependencies {
    implementation("io.ktor:ktor-server-core:$ktor")
    implementation("io.ktor:ktor-server-netty:$ktor")
    implementation("io.ktor:ktor-server-content-negotiation:$ktor")
    implementation("io.ktor:ktor-serialization-kotlinx-json:$ktor")
    implementation("io.ktor:ktor-server-status-pages:$ktor")
    implementation("ch.qos.logback:logback-classic:1.5.18")

    testImplementation(kotlin("test"))
    testImplementation("io.ktor:ktor-server-test-host:$ktor")
}

kotlin { jvmToolchain(21) }

application { mainClass.set("jukebox.MainKt") }

tasks.test { useJUnitPlatform() }
