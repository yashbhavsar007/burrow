package commands

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	pkgs "github.com/hyperledger/burrow/deploy"
	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/proposals"
	"github.com/hyperledger/burrow/deploy/util"
	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

func Deploy(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		chainUrlOpt := cmd.StringOpt("u chain-url", "127.0.0.1:10997", "chain-url to be used in IP:PORT format")

		signerOpt := cmd.StringOpt("s keys", "",
			"IP:PORT of Burrow GRPC service which jobs should or otherwise transaction submitted unsigned for mempool signing in Burrow")

		mempoolSigningOpt := cmd.BoolOpt("p mempool-signing", false,
			"Use Burrow's own keys connection to sign transactions - means that Burrow instance must have access to input account keys. "+
				"Sequence numbers are set as transactions enter the mempool so concurrent transactions can be sent from same inputs.")

		pathOpt := cmd.StringOpt("i dir", "", "root directory of app (will use pwd by default)")

		defaultOutputOpt := cmd.StringOpt("o output", def.DefaultOutputFile,
			"filename for jobs output file. by default, this name will reflect the name passed in on the optional [--file]")

		yamlPathOpt := cmd.StringOpt("f file", "deploy.yaml",
			"path to package file which jobs should use. if also using the --dir flag, give the relative path to jobs file, which should be in the same directory")

		defaultSetsOpt := cmd.StringsOpt("e set", []string{},
			"default sets to use; operates the same way as the [set] jobs, only before the jobs file is ran (and after default address")

		binPathOpt := cmd.StringOpt("b bin-path", "[dir]/bin",
			"path to the bin directory jobs should use when saving binaries after the compile process defaults to --dir + /bin")

		defaultGasOpt := cmd.StringOpt("g gas", "1111111111",
			"default gas to use; can be overridden for any single job")

		jobsOpt := cmd.IntOpt("j jobs", 2,
			"default number of concurrent solidity compilers to run")

		addressOpt := cmd.StringOpt("a address", "",
			"default address to use; operates the same way as the [account] job, only before the deploy file is ran")

		defaultFeeOpt := cmd.StringOpt("n fee", "9999", "default fee to use")

		defaultAmountOpt := cmd.StringOpt("m amount", "9999",
			"default amount to use")

		verboseOpt := cmd.BoolOpt("v verbose", false, "verbose output")

		debugOpt := cmd.BoolOpt("d debug", false, "debug level output")

		proposalVerify := cmd.BoolOpt("proposal-verify", false, "Verify any proposal, do NOT create new proposal or vote")

		proposalVote := cmd.BoolOpt("proposal-vote", false, "Vote for proposal, do NOT create new proposal")

		proposalCreate := cmd.BoolOpt("proposal-create", false, "Create new proposal")

		timeoutOpt := cmd.IntOpt("t timeout", 10, "Timeout to talk to the chain")

		proposalList := cmd.StringOpt("list-proposals state", "", "List proposals, either all, executed, expired, or current")

		cmd.Action = func() {
			do := new(def.DeployArgs)

			if *proposalVerify && *proposalVote {
				output.Fatalf("Cannot combine --proposal-verify and --proposal-vote")
			}

			for _, e := range *defaultSetsOpt {
				s := strings.Split(e, "=")
				if len(s) != 2 || s[0] == "" {
					output.Fatalf("`--set %s' should have format VARIABLE=value", e)
				}
			}

			do.Path = *pathOpt
			do.DefaultOutput = *defaultOutputOpt
			do.YAMLPath = *yamlPathOpt
			do.DefaultSets = *defaultSetsOpt
			do.BinPath = *binPathOpt
			do.DefaultGas = *defaultGasOpt
			do.Address = *addressOpt
			do.DefaultFee = *defaultFeeOpt
			do.DefaultAmount = *defaultAmountOpt
			do.Verbose = *verboseOpt
			do.Debug = *debugOpt
			do.Jobs = *jobsOpt
			do.ProposeVerify = *proposalVerify
			do.ProposeVote = *proposalVote
			do.ProposeCreate = *proposalCreate
			log.SetFormatter(new(PlainFormatter))
			log.SetLevel(log.WarnLevel)
			if do.Verbose {
				log.SetLevel(log.InfoLevel)
			} else if do.Debug {
				log.SetLevel(log.DebugLevel)
			}
			client := def.NewClient(*chainUrlOpt, *signerOpt, *mempoolSigningOpt, time.Duration(*timeoutOpt)*time.Second)
			handleTerm()

			if *proposalList != "" {
				state, err := proposals.ProposalStateFromString(*proposalList)
				if err != nil {
					output.Fatalf(err.Error())
				}
				proposals.ListProposals(client, state)
			} else {
				util.IfExit(pkgs.RunPackage(do, client))
			}
		}
	}
}

type PlainFormatter struct{}

func (f *PlainFormatter) Format(entry *log.Entry) ([]byte, error) {
	var b *bytes.Buffer
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	f.appendMessage(b, entry.Message)
	for _, key := range keys {
		f.appendMessageData(b, key, entry.Data[key])
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *PlainFormatter) appendMessage(b *bytes.Buffer, message string) {
	fmt.Fprintf(b, "%-44s", message)
}

func (f *PlainFormatter) appendMessageData(b *bytes.Buffer, key string, value interface{}) {
	switch key {
	case "":
		b.WriteString("=> ")
	case "=>":
		b.WriteString(key)
		b.WriteByte(' ')
	default:
		b.WriteString(key)
		b.WriteString(" => ")
	}
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}
	b.WriteString(stringVal)
	b.WriteString(" ")
}
