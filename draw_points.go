package main

import (
    "bufio"
	"context"
    "fmt"
	"os"
	"strconv"
	"strings"

    "github.com/golang/geo/r3"
    vizClient "github.com/viam-labs/motion-tools/client/client"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/utils/rpc"
	"go.viam.com/rdk/components/arm"
    "go.viam.com/rdk/services/motion"
    "go.viam.com/rdk/referenceframe"
    "go.viam.com/rdk/spatialmath"
    "go.viam.com/rdk/motionplan"
)

// Viam machine info
var VIAM_MACHINE_IP_ADDRESS = "your string here"
var API_KEY_ID = "your string here"
var API_KEY = "your string here"
var ARM_RESOURCE_NAME = "your string here"

// Text info
var POSE_COLOR = "#EF5350"
var TEXT_SIZE_MULTIPLIER float64 = 45

// End effector position when pen touching paper where you want to start writing from
var END_EFFECTOR_X_WHEN_PEN_TOUCHING_PAPER float64 = 53.81019226834298
var END_EFFECTOR_Y_WHEN_PEN_TOUCHING_PAPER float64 = 261.357484082354 
var END_EFFECTOR_Z_WHEN_PEN_TOUCHING_PAPER float64 = 57.03686307529803
var ADDITIONAL_BUFFER_Z float64 = 20 // The initial position will be this many mm above the paper
var END_EFFECTOR_OV_X_WHEN_PEN_TOUCHING_PAPER float64 = 0.348899278145762
var END_EFFECTOR_OV_Y_WHEN_PEN_TOUCHING_PAPER float64 = 0.0000407496998468827
var END_EFFECTOR_OV_Z_WHEN_PEN_TOUCHING_PAPER float64 = -0.937160228407803
var END_EFFECTOR_OV_THETA_WHEN_PEN_TOUCHING_PAPER float64 = 0.007265004379393861

func main() {
    vizClient.RemoveAllSpatialObjects()

	var end_effector_poses []spatialmath.Pose
	var end_effector_colors []string

	file, err := os.Open("./generate_points/points3d.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var points [][]float64
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		if lineNum == 0 {
			// Skip the header
			lineNum++
			continue
		}

		fields := strings.Split(line, ",")

        if (strings.HasPrefix(line, "#")) {
            fmt.Printf("Skipping visual indicator line: %s\n", line)
            continue
        }

		if len(fields) != 3 {
            fmt.Printf("Skipping malformed line: %s\n", line)
			continue
		}

		point := make([]float64, 3)
		for i, field := range fields {
			val, err := strconv.ParseFloat(strings.TrimSpace(field), 64)
			if err != nil {
				fmt.Printf("Error parsing float: %v\n", err)
				continue
			}
			point[i] = val * TEXT_SIZE_MULTIPLIER
		}

		points = append(points, point)
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	// Construct and print a Pose for each point
	for _, pt := range points {
		x, y, z := pt[0], pt[1], pt[2]

		end_effector_pose := spatialmath.NewPose(
			r3.Vector{X: x + END_EFFECTOR_X_WHEN_PEN_TOUCHING_PAPER, Y: y + END_EFFECTOR_Y_WHEN_PEN_TOUCHING_PAPER, Z: z + END_EFFECTOR_Z_WHEN_PEN_TOUCHING_PAPER},
			&spatialmath.OrientationVectorDegrees{
				OX: END_EFFECTOR_OV_X_WHEN_PEN_TOUCHING_PAPER,
				OY: END_EFFECTOR_OV_Y_WHEN_PEN_TOUCHING_PAPER,
				OZ: END_EFFECTOR_OV_Z_WHEN_PEN_TOUCHING_PAPER,
				Theta: END_EFFECTOR_OV_THETA_WHEN_PEN_TOUCHING_PAPER,
			},
		)

		end_effector_poses = append(end_effector_poses, end_effector_pose)
		end_effector_colors = append(end_effector_colors, POSE_COLOR)
	}

    // Send the poses to the visualizer
	err = vizClient.DrawPoses(end_effector_poses, end_effector_colors, true)
	if err != nil {
        panic(err)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Done generating poses! Look at the visualizer output. Do you want to continue to move the arm to the end_effector_poses? (y/n): ")
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		os.Exit(1)
	}

	// Normalize input
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "y" || input == "yes" {
		fmt.Println("Continuing...")
		// Add your logic here
	} else {
		fmt.Println("Exiting.")
		os.Exit(0)
	}

    // Connect to the Viam machine
    logger := logging.NewDebugLogger("client")
    machine, err := client.New(
        context.Background(),
        VIAM_MACHINE_IP_ADDRESS,
        logger,
        client.WithDialOptions(rpc.WithEntityCredentials( 
            API_KEY_ID,
            rpc.Credentials{
                Type:    rpc.CredentialsTypeAPIKey, 
                Payload: API_KEY,
            })),
    )
    if err != nil {
        logger.Fatal(err)
    }
    defer machine.Close(context.Background())
    
    // Get builtin motion service
    motionService, err := motion.FromRobot(machine, "builtin")
    if err != nil {
       logger.Fatal(err)
    }

    // Move arm to the initial pose.
    fmt.Print("Trying to move to initial pose: ")
    initialPose := spatialmath.NewPose(
        r3.Vector{X: END_EFFECTOR_X_WHEN_PEN_TOUCHING_PAPER, Y: END_EFFECTOR_Y_WHEN_PEN_TOUCHING_PAPER, Z: END_EFFECTOR_Z_WHEN_PEN_TOUCHING_PAPER + ADDITIONAL_BUFFER_Z},
        &spatialmath.OrientationVectorDegrees{
            OX: END_EFFECTOR_OV_X_WHEN_PEN_TOUCHING_PAPER,
            OY: END_EFFECTOR_OV_Y_WHEN_PEN_TOUCHING_PAPER,
            OZ: END_EFFECTOR_OV_Z_WHEN_PEN_TOUCHING_PAPER,
            Theta: END_EFFECTOR_OV_THETA_WHEN_PEN_TOUCHING_PAPER,
        },
    )
    initialPoseInFrame := referenceframe.NewPoseInFrame("world", initialPose)
    moved, err := motionService.Move(context.Background(), motion.MoveReq{
       ComponentName: arm.Named(ARM_RESOURCE_NAME),
       Destination:   initialPoseInFrame,
    })
    if err != nil {
       logger.Fatal(err)
    }
    logger.Info("moved", moved)

   myConstraints := &motionplan.Constraints{
      OrientationConstraint: []motionplan.OrientationConstraint{{1.0}},
   }

    // Move arm to the poses for tracing out the word
	for i, pose := range end_effector_poses {
        fmt.Print("Trying to move to pose: ")
		fmt.Print(pose)
        fmt.Println()

        destinationPoseInFrame := referenceframe.NewPoseInFrame("world", end_effector_poses[i])

        moved, err := motionService.Move(context.Background(), motion.MoveReq{
           ComponentName: arm.Named(ARM_RESOURCE_NAME),
           Destination:   destinationPoseInFrame,
           Constraints: myConstraints,
        })
        if err != nil {
           logger.Fatal(err)
        }
        logger.Info("moved", moved)
	}
}
