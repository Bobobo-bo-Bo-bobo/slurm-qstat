package main

const name = "slurm-qstat"
const version = "1.4.0-20210619"

const versionText = `%s version %s
Copyright (C) 2021 by Andreas Maus <maus@ypbind.de>
This program comes with ABSOLUTELY NO WARRANTY.

pkidb is distributed under the Terms of the GNU General
Public License Version 3. (http://www.gnu.org/copyleft/gpl.html)

Build with go version: %s
`

const helpText = `Usage: %s [--brief] [--filter=<part>,...] [--help] --jobs=<filter>|--nodes|--partitions|--reservations [--sort=<sort>] [--version]

    --brief                     Show brief output

    --filter=<part>,...         Limit output to comma separated list of partitions

    --help                      Show this help text

    --jobs=<filter>             Show job information. <filter> can be one of:
                                    all         - show all jobs
                                    not-running - show not running only (state other than RUNNING)
                                    running     - show only running jobs (state RUNNING)

    --nodes                     Show node information

    --partitions                Show partition information

    --reservations              Show reservation information

    --sort=<sort>               Sort output by field <sort> in ascending order
                                    <sort> is a comma separated list of <object>:<field>
                                    <object> can be prefixed by a minus sign to reverse the sort order of the field
                                    <object> can be one of:
                                        jobs - sort jobs
                                        nodes - sort nodes
                                        partitions - sort partitions
                                        reservations - sort reservations

                                    <field> depends of the <object> type:
                                        jobs:
                                            batchhost - sort by batch host
                                            cpus - sort by cpus
                                            gres - sort by GRES
                                            jobid - sort by job id (this is the default)
                                            licenses - sort by licenses
                                            name - sort by name
                                            nodes - sort by nodes
                                            partition - sort by partitions
                                            reason - sort by state reason
                                            starttime - sort by starttime
                                            state - sort by state
                                            tres - sort by TRES
                                            user - sort by user

                                        nodes:
                                            boards - sort by number of boards
                                            hostname - sort by hostname
                                            nodename - sort by node name (this is the default)
                                            partition - sort by partitions
                                            reason - sort by state reason
                                            slurmversion - sort by reported SLURM version
                                            sockets - sort by number of sockets
                                            state - sort by state
                                            threadsbycore - sort by threads per core
                                            tresallocated - sort by allocated TRES
                                            tresconfigured - sort by configured TRES

                                        partitions:

                                        reservations:

    --version                   Show version information
`

const sortReverse uint8 = 0x80
const maskSortReverse uint8 = 0x7f

const sortNodesMask uint32 = 0x000000ff
const (
	sortNodesByNodeName uint8 = iota
	sortNodesByHostName
	sortNodesByPartition
	sortNodesByState
	sortNodesBySlurmVersion
	sortNodesByTresConfigured
	sortNodesByTresAllocated
	sortNodesBySockets
	sortNodesByBoards
	sortNodesByThreadsPerCore
	sortNodesByReason
)

const sortJobsMask uint32 = 0x0000ff00
const (
	sortJobsByJobID uint8 = iota
	sortJobsByPartition
	sortJobsByUser
	sortJobsByState
	sortJobsByReason
	sortJobsByBatchHost
	sortJobsByNodes
	sortJobsByCPUs
	sortJobsByLicenses
	sortJobsByGres
	sortJobsByTres
	sortJobsByName
	sortJobsByStartTime
)
