/**
 * @file example.hpp
 * @brief Example C++ header file for testing the parser
 */

#ifndef EXAMPLE_HPP
#define EXAMPLE_HPP

#include <vector>
#include <string>
#include <memory>

/**
 * @brief Graphics rendering namespace
 * @details Contains all graphics-related classes and functions
 */
namespace Graphics {
    
    /**
     * @brief Forward declaration of Scene class
     */
    class Scene;
    
    /**
     * @brief Rendering backend enumeration
     */
    enum class Backend {
        OpenGL,     ///< OpenGL backend
        Vulkan,     ///< Vulkan backend
        DirectX12   ///< DirectX 12 backend
    };
    
    /**
     * @brief Main renderer class
     * @details Handles all rendering operations for the graphics engine
     */
    class Renderer {
    public:
        /**
         * @brief Default constructor
         */
        Renderer();
        
        /**
         * @brief Constructor with backend specification
         * @param backend The rendering backend to use
         */
        explicit Renderer(Backend backend);
        
        /**
         * @brief Destructor
         */
        ~Renderer();
        
        /**
         * @brief Initialize the renderer
         * @return true if initialization successful, false otherwise
         * @throws std::runtime_error if backend initialization fails
         */
        bool initialize();
        
        /**
         * @brief Render a scene
         * @param scene The scene to render
         * @param deltaTime Time since last frame in seconds
         */
        void render(const Scene& scene, float deltaTime);
        
        /**
         * @brief Set the rendering backend
         * @param backend The new backend to use
         */
        void setBackend(Backend backend);
        
        /**
         * @brief Get the current backend
         * @return The current rendering backend
         */
        Backend getBackend() const;
        
        /**
         * @brief Check if renderer is initialized
         * @return true if initialized, false otherwise
         */
        bool isInitialized() const { return initialized_; }
        
    protected:
        /**
         * @brief Internal initialization for specific backend
         * @param backend The backend to initialize
         * @return true if successful
         */
        virtual bool initializeBackend(Backend backend);
        
    private:
        Backend backend_;           ///< Current rendering backend
        bool initialized_;          ///< Initialization status
        
        /**
         * @brief Internal render implementation
         * @param scene Scene to render
         */
        void renderInternal(const Scene& scene);
    };
    
    /**
     * @brief Specialized renderer for 2D graphics
     */
    class Renderer2D : public Renderer {
    public:
        /**
         * @brief Constructor for 2D renderer
         */
        Renderer2D();
        
        // Undocumented function to test parser
        void drawSprite(const std::string& texture, float x, float y);
        
    protected:
        bool initializeBackend(Backend backend) override;
    };
    
    // Global functions
    
    /**
     * @brief Create a default renderer instance
     * @return Unique pointer to renderer
     */
    std::unique_ptr<Renderer> createRenderer();
    
    // Undocumented global function
    void shutdownGraphics();
    
    /**
     * @brief Graphics utility functions
     */
    namespace Utils {
        /**
         * @brief Convert color from RGB to HSV
         * @param r Red component (0-255)
         * @param g Green component (0-255)
         * @param b Blue component (0-255)
         * @return HSV values as vector [h, s, v]
         */
        std::vector<float> rgbToHsv(int r, int g, int b);
        
        // Template function (undocumented)
        template<typename T>
        T clamp(T value, T min, T max);
    }
    
} // namespace Graphics

// Global namespace entities

/**
 * @brief Application configuration structure
 */
struct AppConfig {
    std::string title;          ///< Application title
    int width = 800;           ///< Window width
    int height = 600;          ///< Window height
    bool fullscreen = false;   ///< Fullscreen mode
};

/**
 * @brief Application entry point
 * @param config Application configuration
 * @return Exit code
 */
int runApplication(const AppConfig& config);

// Undocumented global variable
extern bool g_debugMode;

/**
 * @brief Convenience typedef for renderer pointer
 */
typedef std::unique_ptr<Graphics::Renderer> RendererPtr;

/**
 * @brief Modern using declaration
 */
using StringVector = std::vector<std::string>;

#endif // EXAMPLE_HPP